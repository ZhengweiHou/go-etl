package loader

import (
	"fmt"
	"sync"

	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/Breeze0806/go-etl/datax/common/spi"
	"github.com/Breeze0806/go-etl/datax/common/spi/reader"
	"github.com/Breeze0806/go-etl/datax/common/spi/writer"
)

var _centor = &centor{
	readers: make(map[string]spi.Reader),
	writers: make(map[string]spi.Writer),
}

func RegisterReader(name string, reader spi.Reader) {
	if err := _centor.registerReader(name, reader); err != nil {
		panic(err)
	}
}

func RegisterWriter(name string, writer spi.Writer) {
	if err := _centor.registerWriter(name, writer); err != nil {
		panic(err)
	}
}

//LoadJobPlugin 目前未正常实现该函数，仅仅是个架子
func LoadJobPlugin(typ plugin.Type, name string) (plugin.Job, error) {
	return &defaultJobPlugin{}, nil
}

func LoadReaderJob(name string) (reader.Job, bool) {
	r, ok := _centor.reader(name)
	if !ok {
		return nil, false
	}
	return r.Job(), true
}

func LoadReaderTask(name string) (reader.Task, bool) {
	r, ok := _centor.reader(name)
	if !ok {
		return nil, false
	}
	return r.Task(), true
}

func LoadWriterJob(name string) (writer.Job, bool) {
	w, ok := _centor.writer(name)
	if !ok {
		return nil, false
	}
	return w.Job(), true
}

func LoadWriterTask(name string) (writer.Task, bool) {
	w, ok := _centor.writer(name)
	if !ok {
		return nil, false
	}
	return w.Task(), true
}

type centor struct {
	readersMu sync.RWMutex
	readers   map[string]spi.Reader

	writersMu sync.RWMutex
	writers   map[string]spi.Writer
}

func (l *centor) registerReader(name string, reader spi.Reader) error {

	l.readersMu.Lock()
	defer l.readersMu.Unlock()
	if reader.Task() == nil || reader.Job() == nil {
		return fmt.Errorf("datax: reader %v has nil job or task", name)
	}

	if _, ok := l.readers[name]; ok {
		return fmt.Errorf("datax: reader %v has already registered", name)
	}

	l.readers[name] = reader
	return nil
}

func (l *centor) reader(name string) (reader spi.Reader, ok bool) {
	l.readersMu.RLock()
	defer l.readersMu.RUnlock()
	reader, ok = l.readers[name]
	return
}

func (l *centor) registerWriter(name string, writer spi.Writer) error {
	l.writersMu.Lock()
	defer l.writersMu.Unlock()
	if writer.Task() == nil || writer.Job() == nil {
		return fmt.Errorf("datax: writer %v has nil job or task", name)
	}
	if _, ok := l.writers[name]; ok {
		return fmt.Errorf("datax: writer %v has already registered", name)
	}
	l.writers[name] = writer
	return nil
}

func (l *centor) writer(name string) (writer spi.Writer, ok bool) {
	l.writersMu.RLock()
	defer l.writersMu.RUnlock()
	writer, ok = l.writers[name]
	return
}

type defaultJobPlugin struct {
	*plugin.BaseJob
}

func (d *defaultJobPlugin) Init() error {
	return nil
}

func (d *defaultJobPlugin) Destroy() error {
	return nil
}
