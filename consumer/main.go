package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Создаем наш контейнер для кэша и сразу заполняем его заказами из БД
	cache := New(5*time.Minute, 10*time.Minute)
	orderList := resultQuery()
	for i := 0; i < len(orderList); i++ {
		cache.Set(orderList[i].OrderUid, orderList[i], 5*time.Minute)
	}

	// Создаем соединение с NATS Streaming Server
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(nats.DefaultURL))
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS Streaming: %v", err)
	}
	defer sc.Close()

	// Подписываемся на канал и обрабатываем получаемые сообщения
	sub, err := sc.Subscribe(channel, func(msg *stan.Msg) {
		log.Printf("Получено сообщение: %s\n", string(msg.Data))
		order, flg := jsonCheck(msg.Data)
		if flg {
			runQuery(fromOrderToQuery(order))
			cache.Set(order.OrderUid, order, 5*time.Minute)
			fmt.Println(cache)
		}
	}, stan.DurableName("durable-subscriber"))
	if err != nil {
		log.Fatalf("Ошибка подписки на канал: %v", err)
	}
	defer sub.Unsubscribe()

	// Тут разворачиваем http-сервер

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html-package/index.html")
	})
	http.HandleFunc("/getOrder", func(w http.ResponseWriter, r *http.Request) {
		idOrder := r.FormValue("idOrder")
		o, flg := cache.Get(idOrder) //b563feb7b2b84b6test2
		if !flg {
			o = Order{}
		}
		order := o.(Order)
		tmpl, _ := template.ParseFiles("html-package/order.html")
		tmpl.Execute(w, order)
		//idOrder := r.FormValue("idOrder")
		//fmt.Fprintf(w, "ID заказа: %s", idOrder)
	})
	http.ListenAndServe(":3000", nil)

	// Ожидаем сигнал для завершения программы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Завершение программы")
}
