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

package pushing

import (
	"github.com/mutalisk999/gitbitex-service-group/conf"
	"github.com/mutalisk999/gitbitex-service-group/matching"
	"github.com/mutalisk999/gitbitex-service-group/service"
	"github.com/siddontang/go-log/log"
	"sync"
	"time"
)

var productsSupported sync.Map

func StartServer() {
	gbeConfig := conf.GetConfig()

	sub := newSubscription()

	newRedisStream(sub).Start()

	go func() {
		for {
			products, err := service.GetProducts()
			if err != nil {
				panic(err)
			}
			for _, product := range products {
				_, ok := productsSupported.Load(product.Id)
				if !ok {
					newTickerStream(product.Id, sub, matching.NewKafkaLogReader("tickerStream", product.Id, gbeConfig.Kafka.Brokers)).Start()
					newMatchStream(product.Id, sub, matching.NewKafkaLogReader("matchStream", product.Id, gbeConfig.Kafka.Brokers)).Start()
					newOrderBookStream(product.Id, sub, matching.NewKafkaLogReader("orderBookStream", product.Id, gbeConfig.Kafka.Brokers)).Start()
					productsSupported.Store(product.Id, true)
					log.Infof("start stream for %s ok", product.Id)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	go NewServer(gbeConfig.PushServer.Addr, gbeConfig.PushServer.Path, sub).Run()

	log.Info("websocket server ok")
}
