// Copyright 2020 the go-etl Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"context"
	"sync"
	"time"

	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/Breeze0806/go-etl/datax/common/spi/writer"
	"github.com/Breeze0806/go-etl/datax/core/transport/exchange"
	"github.com/Breeze0806/go-etl/element"
	"github.com/Breeze0806/go-etl/storage/stream/file"
)

//Task 任务
type Task struct {
	*writer.BaseTask

	streamer  *file.OutStreamer
	conf      Config
	newConfig func(conf *config.JSON) (Config, error)
	content   *config.JSON
}

//NewTask 通过获取配置newConfig创建任务
func NewTask(newConfig func(conf *config.JSON) (Config, error)) *Task {
	return &Task{
		BaseTask:  writer.NewBaseTask(),
		newConfig: newConfig,
	}
}

//Init 初始化
func (t *Task) Init(ctx context.Context) (err error) {
	var name string
	if name, err = t.PluginConf().GetString("creator"); err != nil {
		return t.Wrapf(err, "GetString fail")
	}
	var filename string
	if filename, err = t.PluginJobConf().GetString("path"); err != nil {
		return t.Wrapf(err, "GetString fail")
	}

	if t.content, err = t.PluginJobConf().GetConfig("content"); err != nil {
		return t.Wrapf(err, "GetString fail")
	}

	if t.conf, err = t.newConfig(t.content); err != nil {
		return t.Wrapf(err, "newConfig fail")
	}

	if t.streamer, err = file.NewOutStreamer(name, filename); err != nil {
		return t.Wrapf(err, "NewOutStreamer fail")
	}

	return
}

//Destroy 销毁
func (t *Task) Destroy(ctx context.Context) (err error) {
	if t.streamer != nil {
		err = t.streamer.Close()
	}
	return t.Wrapf(err, "Close fail")
}

//StartWrite 开始写
func (t *Task) StartWrite(ctx context.Context, receiver plugin.RecordReceiver) (err error) {
	var sw file.StreamWriter
	if sw, err = t.streamer.Writer(t.content); err != nil {
		return t.Wrapf(err, "Writer fail")
	}

	recordChan := make(chan element.Record)
	var rerr error
	afterCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	//通过该携程读取记录接受器receiver的记录放入recordChan
	go func() {
		defer func() {
			wg.Done()
			//关闭recordChan
			close(recordChan)
			log.Debugf(t.Format("get records end"))
		}()
		log.Debugf(t.Format("start to get records"))
		for {
			select {
			case <-afterCtx.Done():
				return
			default:
			}
			var record element.Record
			record, rerr = receiver.GetFromReader()
			if rerr != nil && rerr != exchange.ErrEmpty {
				return
			}

			//当记录接受器receiver返回不为空错误，写入recordChan
			if rerr != exchange.ErrEmpty {
				select {
				//防止在ctx关闭时不写入recordChan
				case <-afterCtx.Done():
					return
				case recordChan <- record:
				}

			}
		}
	}()
	ticker := time.NewTicker(t.conf.GetBatchTimeout())
	defer ticker.Stop()
	cnt := 0
	log.Debugf(t.Format("start to write"))
	for {
		select {
		case record, ok := <-recordChan:
			if !ok {
				//当写入结束时，将剩余的记录写入数据库
				if cnt > 0 {
					if err = sw.Flush(); err != nil {
						log.Errorf(t.Format("Flush error: %v"), err)
					}
				}
				if err == nil {
					err = rerr
				}
				goto End
			}

			//写入文件
			if err = sw.Write(record); err != nil {
				log.Errorf(t.Format("Write error: %v"), err)
				goto End
			}
			cnt++
			//当数据量超过单次批量数时 写入文件
			if cnt >= t.conf.GetBatchSize() {
				if err = sw.Flush(); err != nil {
					log.Errorf(t.Format("Flush error: %v"), err)
					goto End
				}
				cnt = 0
			}
		//当写入数据未达到单次批量数，超时也写入
		case <-ticker.C:
			if cnt > 0 {
				if err = sw.Flush(); err != nil {
					log.Errorf(t.Format("Flush error: %v"), err)
					goto End
				}
			}
			cnt = 0
		}
	}
End:
	if cerr := sw.Close(); cerr != nil {
		log.Errorf(t.Format("Close error: %v"), cerr)
	}
	cancel()
	log.Debugf(t.Format("wait all goroutine"))
	//等待携程结束
	wg.Wait()
	log.Debugf(t.Format(" wait all goroutine end"))
	switch {
	//当外部取消时，开始写入不是错误
	case ctx.Err() != nil:
		return nil
	//当错误是停止时，也不是错误
	case err == exchange.ErrTerminate:
		return nil
	}
	return t.Wrapf(err, "")
}
