// Copyright 2019 GitBitEx.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matching

import (
	"github.com/mutalisk999/gitbitex-service-group/conf"
	"github.com/mutalisk999/gitbitex-service-group/service"
	"github.com/siddontang/go-log/log"
	"sync"
	"time"
)

var productsSupported sync.Map

func StartEngine() {
	gbeConfig := conf.GetConfig()

	go func() {
		for {
			products, err := service.GetProducts()
			if err != nil {
				panic(err)
			}
			for _, product := range products {
				_, ok := productsSupported.Load(product.Id)
				if !ok {
					orderReader := NewKafkaOrderReader(product.Id, gbeConfig.Kafka.Brokers)
					snapshotStore := NewRedisSnapshotStore(product.Id)
					logStore := NewKafkaLogStore(product.Id, gbeConfig.Kafka.Brokers)
					matchEngine := NewEngine(product, orderReader, logStore, snapshotStore)
					matchEngine.Start()
					productsSupported.Store(product.Id, true)
					log.Infof("start match engine for %s ok", product.Id)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	log.Info("match engine ok")
}
