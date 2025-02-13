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

package taskgroup

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/common/plugin/loader"
)

func testTaskExecer(ctx context.Context, taskConf *config.JSON, jobID, taskGroupID int64, attemptCount int) *taskExecer {
	t, err := newTaskExecer(ctx, taskConf, jobID, taskGroupID, attemptCount)
	if err != nil {
		panic(err)
	}
	return t
}

func initLoader(name string, errs []error) {
	loader.RegisterReader(name, newMockReader(errs))
	loader.RegisterWriter(name, newMockWriter(errs))
}

func Test_newTaskExecer(t *testing.T) {
	resetLoader()
	initLoader("mock", []error{
		nil, nil, nil, nil, nil,
	})

	type args struct {
		ctx          context.Context
		taskConf     *config.JSON
		jobID        int64
		taskGroupID  int64
		attemptCount int
	}
	tests := []struct {
		name    string
		args    args
		wantT   *taskExecer
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":1,
						"reader":{
							"name":"mock",
							"parameter":{}
						},
						"writer":{
							"name":"mock",
							"parameter":{}
						}
					}`),
				jobID:        1,
				taskGroupID:  1,
				attemptCount: 0,
			},
			wantErr: false,
		},

		{
			name: "2",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":2,
						"reader":{
							"name":"mock2"
						},
						"writer":{
							"name":"mock2"
						}
					}`),
				jobID:        2,
				taskGroupID:  2,
				attemptCount: 0,
			},
			wantErr: true,
		},
		{
			name: "3",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":2,
						"reader":{
							"name":1
						},
						"writer":{
							"name":"mock2"
						}
					}`),
				jobID:        3,
				taskGroupID:  3,
				attemptCount: 0,
			},
			wantErr: true,
		},
		{
			name: "4",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":3,
						"reader":{
							"name":"mock"
						},
						"writer":{
							"name":"mock2"
						}
					}`),
				jobID:        4,
				taskGroupID:  4,
				attemptCount: 0,
			},
			wantErr: true,
		},
		{
			name: "5",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":2,
						"reader":{
							"name":"mock"
						},
						"writer":{
							"name":2
						}
					}`),
				jobID:        5,
				taskGroupID:  5,
				attemptCount: 0,
			},
			wantErr: true,
		},
		{
			name: "6",
			args: args{
				ctx: context.Background(),
				taskConf: testJSONFromString(`{
						"taskId":"6",
						"reader":{
							"name":"mock",
							"parameter":{}
						},
						"writer":{
							"name":"mock",
							"parameter":{}
						}
					}`),
				jobID:        6,
				taskGroupID:  6,
				attemptCount: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTaskExecer(tt.args.ctx, tt.args.taskConf, tt.args.jobID, tt.args.taskGroupID, tt.args.attemptCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("newTaskExecer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_taskExecer_Do(t *testing.T) {
	resetLoader()
	initLoader("mock", []error{
		nil, nil, nil, nil, nil,
	})
	initLoader("mock1", []error{
		errors.New("mock test error"), nil, nil, nil, nil,
	})
	tests := []struct {
		name    string
		t       *taskExecer
		wantErr bool
	}{
		{
			name: "1",
			t: testTaskExecer(context.Background(), testJSONFromString(`{
				"taskId":1,
				"reader":{
					"name":"mock",
					"parameter":{}
				},
				"writer":{
					"name":"mock",
					"parameter":{}
				}
			}`), 1, 1, 0),
			wantErr: false,
		},

		{
			name: "2",
			t: testTaskExecer(context.Background(), testJSONFromString(`{
				"taskId":1,
				"reader":{
					"name":"mock1",
					"parameter":{}
				},
				"writer":{
					"name":"mock1",
					"parameter":{}
				}
			}`), 2, 2, 0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Do()
			if (err != nil) != tt.wantErr {
				t.Errorf("taskExecer.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(err)
		})
	}
}
