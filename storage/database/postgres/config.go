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

package postgres

import (
	"encoding/json"
	"net/url"

	"github.com/Breeze0806/go-etl/config"
)

//Config postgres配置
type Config struct {
	URL      string `json:"url"`      //数据库url，包含数据库地址，数据库其他参数
	Username string `json:"username"` //用户名
	Password string `json:"password"` //密码
}

//NewConfig 创建postgres配置，如果格式不符合要求，就会报错
func NewConfig(conf *config.JSON) (c *Config, err error) {
	c = &Config{}
	err = json.Unmarshal([]byte(conf.String()), c)
	if err != nil {
		return nil, err
	}
	return
}

//FormatDSN 生成数据源连接信息，url有错会报错
func (c *Config) FormatDSN() (dsn string, err error) {
	var URL *url.URL
	URL, err = url.Parse(c.URL)
	if err != nil {
		return
	}

	URL.User = url.User(c.Username)
	if c.Password != "" {
		URL.User = url.UserPassword(c.Username, c.Password)
	}

	return URL.String(), nil
}
