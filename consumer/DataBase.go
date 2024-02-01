package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func resultQuery() []Order {
	connStr := fmt.Sprintf("user=%v password=%v dbname=%v sslmode=%v", dbUser, dbPass, dbName, dbMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Query("SELECT\no.order_uid,\no.track_number,\no.entry,\no.locale,\no.internal_signature,\no.customer_id,\no.delivery_service,\no.shardkey,\no.sm_id,\no.date_created,\no.oof_shard,\nd.name,\nd.phone,\nd.zip,\nd.city,\nd.address,\nd.region,\nd.email,\npa.transaction,\npa.request_id,\npa.currency,\npa.provider,\npa.amount::numeric::float8::int,\nEXTRACT(EPOCH FROM pa.payment_dt)::integer as payment_dt,\npa.bank,\npa.delivery_cost::numeric::float8::int,\npa.goods_total::numeric::float8::int,\npa.custom_fee::numeric::float8::int\nFROM public.\"ORDERS\" o left join public.\"DELIVERY\" d on  d.id  = o.delivery_id left join public.\"PAYMENTS\" pa on pa.id = o.payments_id")
	if err != nil {
		panic(err)
	}
	defer result.Close()

	var orders []Order
	for result.Next() {
		o := Order{}
		err := result.Scan(
			&o.OrderUid,
			&o.TrackNumber,
			&o.Entry,
			&o.Locale,
			&o.InternalSignature,
			&o.CustomerId,
			&o.DeliveryService,
			&o.Shardkey,
			&o.SmId,
			&o.DateCreated,
			&o.OofShard,
			&o.Delivery.Name,
			&o.Delivery.Phone,
			&o.Delivery.Zip,
			&o.Delivery.City,
			&o.Delivery.Address,
			&o.Delivery.Region,
			&o.Delivery.Email,
			&o.Payment.Transaction,
			&o.Payment.RequestId,
			&o.Payment.Currency,
			&o.Payment.Provider,
			&o.Payment.Amount,
			&o.Payment.PaymentDt,
			&o.Payment.Bank,
			&o.Payment.DeliveryCost,
			&o.Payment.GoodsTotal,
			&o.Payment.CustomFee)
		if err != nil {
			fmt.Println(err)
			continue
		}

		iResult, err := db.Query(fmt.Sprintf("select\ni.chrt_id,\ni.track_number,\ni.price::numeric::float8::int,\ni.rid,\ni.name,\ni.sale,\ni.size,\ni.total_price::numeric::float8::int,\ni.nm_id,\ni.brand,\ni.status\nfrom public.\"ITEMS\" i\nwhere i.track_number = '%v'", o.TrackNumber))
		if err != nil {
			panic(err)
		}
		for iResult.Next() {
			item := ItemFromOrder{}
			err := iResult.Scan(
				&item.ChrtId,
				&item.TrackNumber,
				&item.Price,
				&item.Rid,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.TotalPrice,
				&item.NmId,
				&item.Brand,
				&item.Status)
			if err != nil {
				fmt.Println(err)
				continue
			}
			o.Items = append(o.Items, item)
		}
		iResult.Close()
		orders = append(orders, o)
	}
	return orders
}

func runQuery(s string) {
	connStr := "user=muzalevsky-n-a password=WBIntern dbname=WildBerriesLevel0 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.LastInsertId()) // не поддерживается
	fmt.Println(result.RowsAffected()) // количество добавленных строк
}

func query() string {
	return "SELECT * FROM public.\"ORDERS\""
}

func fromOrderToQuery(order Order) string {
	var DelQuery = fmt.Sprintf("INSERT INTO public.\"DELIVERY\" (name,phone,zip,city,address,region,email) VALUES ('%v','%v','%v','%v','%v','%v','%v');\n",
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email)
	//fmt.Println("Такой запрос на добавление в DELIVERY получился:\n", DelQuery)
	var PayQuery = fmt.Sprintf("INSERT INTO public.\"PAYMENTS\" (transaction,request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee) VALUES ('%v','%v','%v','%v',%v,to_timestamp(%v),'%v',%v,%v,%v);\n",
		order.Payment.Transaction,
		order.Payment.RequestId,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee)
	var OrdQuery = fmt.Sprintf("INSERT INTO public.\"ORDERS\" (order_uid,track_number,entry,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard,delivery_id,payments_id) VALUES ('%v','%v','%v','%v','%v','%v','%v','%v',%v,'%v','%v',",
		order.OrderUid,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerId,
		order.DeliveryService,
		order.Shardkey,
		order.SmId,
		order.DateCreated,
		order.OofShard) +
		fmt.Sprintf("(select max(id) from public.\"DELIVERY\" where name = '%v' and phone = '%v' and zip = '%v' and city = '%v' and address = '%v' and region = '%v' and email = '%v'),",
			order.Delivery.Name,
			order.Delivery.Phone,
			order.Delivery.Zip,
			order.Delivery.City,
			order.Delivery.Address,
			order.Delivery.Region,
			order.Delivery.Email) +
		fmt.Sprintf("(select max(id) from public.\"PAYMENTS\" where transaction = '%v' and request_id = '%v' and currency = '%v' and provider = '%v' and amount = %v::money and payment_dt = to_timestamp(%v) and bank = '%v' and delivery_cost = %v::money and goods_total = %v::money and custom_fee = %v::money));\n",
			order.Payment.Transaction,
			order.Payment.RequestId,
			order.Payment.Currency,
			order.Payment.Provider,
			order.Payment.Amount,
			order.Payment.PaymentDt,
			order.Payment.Bank,
			order.Payment.DeliveryCost,
			order.Payment.GoodsTotal,
			order.Payment.CustomFee)
	var ItmQuery = ""
	for i := 0; i < len(order.Items); i++ {
		ItmQuery += fmt.Sprintf("INSERT INTO public.\"ITEMS\" (chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status) VALUES (%v,'%v',%v,'%v','%v',%v,'%v',%v,%v,'%v',%v);\n",
			order.Items[i].ChrtId,
			order.Items[i].TrackNumber,
			order.Items[i].Price,
			order.Items[i].Rid,
			order.Items[i].Name,
			order.Items[i].Sale,
			order.Items[i].Size,
			order.Items[i].TotalPrice,
			order.Items[i].NmId,
			order.Items[i].Brand,
			order.Items[i].Status)
	}
	fmt.Println(DelQuery + PayQuery + OrdQuery + ItmQuery)
	return DelQuery + PayQuery + OrdQuery + ItmQuery
}
