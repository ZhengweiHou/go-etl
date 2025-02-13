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

package writer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/Breeze0806/go-etl/datax/common/spi/writer"
)

type mockJob struct {
	*plugin.BaseJob
}

func (m *mockJob) Init(ctx context.Context) (err error) {
	return
}

func (m *mockJob) Destroy(ctx context.Context) (err error) {
	return
}

func (m *mockJob) Split(ctx context.Context, number int) ([]*config.JSON, error) {
	return nil, nil
}

type mockTask struct {
	*writer.BaseTask
}

func (m *mockTask) Init(ctx context.Context) (err error) {
	return
}

func (m *mockTask) Destroy(ctx context.Context) (err error) {
	return
}

func (m *mockTask) StartWrite(ctx context.Context, receiver plugin.RecordReceiver) (err error) {
	return
}

type mockWriter struct {
	pluginConf *config.JSON
}

func newMockWriter(filename string) (w *mockWriter, err error) {
	w = &mockWriter{}
	w.pluginConf, err = config.NewJSONFromFile(filename)
	if err != nil {
		return nil, err
	}
	return
}

func (w *mockWriter) ResourcesConfig() *config.JSON {
	return w.pluginConf
}

func (w *mockWriter) Job() writer.Job {
	return &mockJob{}
}

func (w *mockWriter) Task() writer.Task {
	return &mockTask{}
}

type mockWriterMaker struct {
	err error
}

func (m *mockWriterMaker) FromFile(path string) (Writer, error) {
	return newMockWriter(path)
}

func (m *mockWriterMaker) Default() (Writer, error) {
	return nil, nil
}

type mockWriterMaker2 struct {
	path string
	err  error
}

func (m *mockWriterMaker2) FromFile(path string) (Writer, error) {
	m.path = path + ".tmp"
	return newMockWriter(path)
}

func (m *mockWriterMaker2) Default() (Writer, error) {
	r, err := newMockWriter(m.path)
	r.pluginConf.Set("name", "reader2")
	return r, err
}

func TestRegisterWriter(t *testing.T) {
	type args struct {
		maker   Maker
		prepare func()
		post    func()
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				maker:   &mockWriterMaker{},
				prepare: func() {},
				post:    func() {},
			},
			want: filepath.Join("github.com", "Breeze0806", "go-etl", "datax", "plugin", "writer", "resources", "plugin.json"),
		},
		{
			name: "2",
			args: args{
				maker: &mockWriterMaker2{},
				prepare: func() {
					f := filepath.Join("resources", "plugin.json")
					os.Rename(f, f+".tmp")
				},
				post: func() {
					f := filepath.Join("resources", "plugin.json")
					os.Rename(f+".tmp", f)
				},
			},
			want: filepath.Join("github.com", "Breeze0806", "go-etl", "datax", "plugin", "writer", "resources", "plugin.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.prepare()
			defer tt.args.post()
			got, err := RegisterWriter(tt.args.maker)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(got, tt.want) {
				t.Errorf("RegisterWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}
