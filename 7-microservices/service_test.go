package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"reflect"
	"runtime"
	"stepikGoWebServices/generated/service"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	// какой адрес-порт слушать серверу
	listenAddr string = "127.0.0.1:8082"

	// кого по каким методам пускать
	ACLData string = `{
	"logger1":          ["/service.Admin/Logging"],
	"logger2":          ["/service.Admin/Logging"],
	"stat1":            ["/service.Admin/Statistics"],
	"stat2":            ["/service.Admin/Statistics"],
	"business_user":    ["/service.BusinessLogic/Check", "/service.BusinessLogic/Add"],
	"business_admin":   ["/service.BusinessLogic/*"],
	"after_disconnect": ["/service.BusinessLogic/Add"]
}`
)

func wait(amount int) {
	time.Sleep(time.Duration(amount) * 10 * time.Millisecond)
}

// утилитарная функция для коннекта к серверу
func getGrpcConn(t *testing.T) *grpc.ClientConn {
	grpcConn, err := grpc.NewClient(
		listenAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("cannot connect to grpc: %v", err)
	}
	return grpcConn
}

// получаем контекст с нужными метаданными для ACL
func getConsumerCtx(consumerName string) context.Context {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
	ctx := context.Background()
	md := metadata.Pairs(
		"consumer", consumerName,
	)
	return metadata.NewOutgoingContext(ctx, md)
}

// получаем контекст с нужными метаданными для ACL с возможностью отмены
func getConsumerCtxWithCancel(consumerName string) (context.Context, context.CancelFunc) {
	ctx, cancelFn := context.WithCancel(context.Background())
	md := metadata.Pairs(
		"consumer", consumerName,
	)
	return metadata.NewOutgoingContext(ctx, md), cancelFn
}

// старт-стоп сервера
func TestServerStartStop(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())

	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	wait(1)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт
	wait(1)

	// теперь проверим что вы освободили порт и мы можем стартовать сервер ещё раз
	ctx, finish = context.WithCancel(context.Background())

	err = StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server again: %v", err)
	}

	wait(1)
	finish()
	wait(1)
}

// у вас наверняка будет что-то выполняться в отдельных горутинах
// этим тестом мы проверяем что вы останавливаете все горутины которые у вас были и нет утечек
// некоторый запас ( goroutinesPerTwoIterations*5 ) остаётся на случай рантайм горутин
func TestServerLeak(t *testing.T) {
	goroutinesStart := runtime.NumGoroutine()
	TestServerStartStop(t)
	goroutinesPerTwoIterations := runtime.NumGoroutine() - goroutinesStart

	goroutinesStart = runtime.NumGoroutine()
	goroutinesStat := make([]int, 0)

	for i := 0; i <= 25; i++ {
		TestServerStartStop(t)
		goroutinesStat = append(goroutinesStat, runtime.NumGoroutine())
	}

	goroutinesPerFiftyIterations := runtime.NumGoroutine() - goroutinesStart

	if goroutinesPerFiftyIterations > goroutinesPerTwoIterations*5 {
		t.Fatalf("looks like you have goroutines leak: %+v", goroutinesStat)
	}
}

// ACL (права на методы доступа) парсится корректно
func TestACLParseError(t *testing.T) {
	err := StartMyMicroservice(context.Background(), listenAddr, "{.;")
	if err == nil {
		t.Fatalf("expected error on bad acl json, have nil")
	}
}

// ACL (права на методы доступа) работает корректно
func TestACL(t *testing.T) {
	wait(1)
	ctx, finish := context.WithCancel(context.Background())

	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	wait(1)
	defer func() {
		finish()
		wait(1)
	}()

	conn := getGrpcConn(t)
	defer conn.Close()

	businessLogic := service.NewBusinessLogicClient(conn)
	admin := service.NewAdminClient(conn)

	for idx, ctx := range []context.Context{
		context.Background(),            // нет поля для ACL
		getConsumerCtx("unknown"),       // поле есть, неизвестный консьюмер
		getConsumerCtx("business_user"), // поле есть, нет доступа
	} {
		_, err = businessLogic.Test(ctx, &service.Nothing{})

		if err == nil {
			t.Fatalf("[%d] ACL fail: expected err on disallowed method", idx)
		} else if code := status.Code(err); code != codes.Unauthenticated {
			t.Fatalf("[%d] ACL fail: expected Unauthenticated code, got %v", idx, code)
		}
	}

	_, err = businessLogic.Check(getConsumerCtx("business_user"), &service.Nothing{})
	if err != nil {
		t.Fatalf("ACL fail: unexpected error: %v", err)
	}

	_, err = businessLogic.Check(getConsumerCtx("business_admin"), &service.Nothing{})
	if err != nil {
		t.Fatalf("ACL fail: unexpected error: %v", err)
	}

	_, err = businessLogic.Test(getConsumerCtx("business_admin"), &service.Nothing{})
	if err != nil {
		t.Fatalf("ACL fail: unexpected error: %v", err)
	}

	logger, err := admin.Logging(getConsumerCtx("unknown"), &service.Nothing{})
	_, err = logger.Recv()

	if err == nil {
		t.Fatalf("ACL fail: expected err on disallowed method")
	} else if code := status.Code(err); code != codes.Unauthenticated {
		t.Fatalf("ACL fail: expected Unauthenticated code, got %v", code)
	}
}

func TestLogging(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())

	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	wait(1)
	defer func() {
		finish()
		wait(1)
	}()

	conn := getGrpcConn(t)
	defer conn.Close()

	businessLogic := service.NewBusinessLogicClient(conn)
	admin := service.NewAdminClient(conn)

	logStream1, err := admin.Logging(getConsumerCtx("logger1"), &service.Nothing{})
	time.Sleep(1 * time.Millisecond)

	logStream2, err := admin.Logging(getConsumerCtx("logger2"), &service.Nothing{})

	logData1 := make([]*service.Event, 0)
	logData2 := make([]*service.Event, 0)

	wait(1)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			fmt.Println("looks like you dont send anything to log stream in 3 sec")
			t.Errorf("looks like you dont send anything to log stream in 3 sec")
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 4; i++ {
			event, err := logStream1.Recv()
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}

			// event.Host читайте как event.RemoteAddr
			if !strings.HasPrefix(event.GetHost(), "127.0.0.1:") || event.GetHost() == listenAddr {
				t.Errorf("bad host: %v", event.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData1 = append(
				logData1,
				&service.Event{Consumer: event.Consumer, Method: event.Method},
			)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			event, err := logStream2.Recv()
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}

			if !strings.HasPrefix(event.GetHost(), "127.0.0.1:") || event.GetHost() == listenAddr {
				t.Errorf("bad host: %v", event.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData2 = append(
				logData2,
				&service.Event{Consumer: event.Consumer, Method: event.Method},
			)
		}
	}()

	businessLogic.Check(getConsumerCtx("business_user"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	businessLogic.Check(getConsumerCtx("business_admin"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	businessLogic.Test(getConsumerCtx("business_admin"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	wg.Wait()

	expectedLogData1 := []*service.Event{
		{Consumer: "logger2", Method: "/service.Admin/Logging"},
		{Consumer: "business_user", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Test"},
	}
	expectedLogData2 := []*service.Event{
		{Consumer: "business_user", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Test"},
	}

	if !reflect.DeepEqual(logData1, expectedLogData1) {
		t.Fatalf("logs1 dont match\nhave %+v\nwant %+v", logData1, expectedLogData1)
	}
	if !reflect.DeepEqual(logData2, expectedLogData2) {
		t.Fatalf("logs2 dont match\nhave %+v\nwant %+v", logData2, expectedLogData2)
	}
}

func TestStat(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())

	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	wait(1)
	defer func() {
		finish()
		wait(2)
	}()

	conn := getGrpcConn(t)
	defer conn.Close()

	businessLogic := service.NewBusinessLogicClient(conn)
	admin := service.NewAdminClient(conn)

	statStream1, err := admin.Statistics(
		getConsumerCtx("stat1"),
		&service.StatInterval{IntervalSeconds: 2},
	)
	wait(1)
	statStream2, err := admin.Statistics(
		getConsumerCtx("stat2"),
		&service.StatInterval{IntervalSeconds: 3},
	)

	mu := &sync.Mutex{}
	stat1 := &service.Stat{}
	stat2 := &service.Stat{}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for {
			stat, err := statStream1.Recv()
			if err != nil && err != io.EOF {
				return
			} else if err == io.EOF {
				break
			}
			mu.Lock()
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			stat1 = &service.Stat{
				ByMethod:   stat.ByMethod,
				ByConsumer: stat.ByConsumer,
			}
			mu.Unlock()
		}
	}()

	go func() {
		for {
			stat, err := statStream2.Recv()
			if err != nil && err != io.EOF {
				return
			} else if err == io.EOF {
				break
			}
			mu.Lock()
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			stat2 = &service.Stat{
				ByMethod:   stat.ByMethod,
				ByConsumer: stat.ByConsumer,
			}
			mu.Unlock()
		}
	}()

	wait(1)

	businessLogic.Check(getConsumerCtx("business_user"), &service.Nothing{})
	businessLogic.Add(getConsumerCtx("business_user"), &service.Nothing{})
	businessLogic.Test(getConsumerCtx("business_admin"), &service.Nothing{})

	wait(200)

	expectedStat1 := &service.Stat{
		ByMethod: map[string]uint64{
			"/service.BusinessLogic/Check": 1,
			"/service.BusinessLogic/Add":   1,
			"/service.BusinessLogic/Test":  1,
			"/service.Admin/Statistics":    1,
		},
		ByConsumer: map[string]uint64{
			"business_user":  2,
			"business_admin": 1,
			"stat2":          1,
		},
	}

	mu.Lock()
	if !reflect.DeepEqual(stat1, expectedStat1) {
		t.Fatalf("stat1-1 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	}
	mu.Unlock()

	businessLogic.Add(getConsumerCtx("business_admin"), &service.Nothing{})

	wait(220)

	expectedStat1 = &service.Stat{
		Timestamp: 0,
		ByMethod: map[string]uint64{
			"/service.BusinessLogic/Add": 1,
		},
		ByConsumer: map[string]uint64{
			"business_admin": 1,
		},
	}
	expectedStat2 := &service.Stat{
		Timestamp: 0,
		ByMethod: map[string]uint64{
			"/service.BusinessLogic/Check": 1,
			"/service.BusinessLogic/Add":   2,
			"/service.BusinessLogic/Test":  1,
		},
		ByConsumer: map[string]uint64{
			"business_user":  2,
			"business_admin": 2,
		},
	}

	mu.Lock()
	if !reflect.DeepEqual(stat1, expectedStat1) {
		t.Fatalf("stat1-2 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	}
	if !reflect.DeepEqual(stat2, expectedStat2) {
		t.Fatalf("stat2 dont match\nhave %+v\nwant %+v", stat2, expectedStat2)
	}
	mu.Unlock()

	finish()
}

func TestWorkAfterDisconnect(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())

	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	wait(1)
	defer func() {
		finish()
		wait(1)
	}()

	conn := getGrpcConn(t)
	defer conn.Close()

	businessLogic := service.NewBusinessLogicClient(conn)
	admin := service.NewAdminClient(conn)

	ctx1, cancel1 := getConsumerCtxWithCancel("logger1")

	logStream1, err := admin.Logging(ctx1, &service.Nothing{})
	time.Sleep(1 * time.Millisecond)

	logStream2, err := admin.Logging(getConsumerCtx("logger2"), &service.Nothing{})

	logData1 := make([]*service.Event, 0)
	logData2 := make([]*service.Event, 0)

	wait(1)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			fmt.Println("looks like you dont send anything to log stream in 3 sec")
			t.Errorf("looks like you dont send anything to log stream in 3 sec")
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 4; i++ {
			event, err := logStream1.Recv()
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}
			// event.Host читайте как event.RemoteAddr
			if !strings.HasPrefix(event.GetHost(), "127.0.0.1:") || event.GetHost() == listenAddr {
				t.Errorf("bad host: %v", event.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData1 = append(
				logData1,
				&service.Event{Consumer: event.Consumer, Method: event.Method},
			)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			event, err := logStream2.Recv()
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}
			if !strings.HasPrefix(event.GetHost(), "127.0.0.1:") || event.GetHost() == listenAddr {
				t.Errorf("bad host: %v", event.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData2 = append(
				logData2,
				&service.Event{Consumer: event.Consumer, Method: event.Method},
			)
		}
	}()

	businessLogic.Check(getConsumerCtx("business_user"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	businessLogic.Check(getConsumerCtx("business_admin"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	businessLogic.Test(getConsumerCtx("business_admin"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	// CHANGED
	wait(12)
	cancel1()
	wait(12)

	businessLogic.Add(getConsumerCtx("after_disconnect"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)

	businessLogic.Add(getConsumerCtx("after_disconnect"), &service.Nothing{})
	time.Sleep(2 * time.Millisecond)
	// END CHANGED

	wg.Wait()

	expectedLogData1 := []*service.Event{
		{Consumer: "logger2", Method: "/service.Admin/Logging"},
		{Consumer: "business_user", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Test"},
	}
	expectedLogData2 := []*service.Event{
		{Consumer: "business_user", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Check"},
		{Consumer: "business_admin", Method: "/service.BusinessLogic/Test"},
		{Consumer: "after_disconnect", Method: "/service.BusinessLogic/Add"}, // CHANGED
		{Consumer: "after_disconnect", Method: "/service.BusinessLogic/Add"}, // CHANGED
	}

	if !reflect.DeepEqual(logData1, expectedLogData1) {
		t.Fatalf("logs1 dont match\nhave %+v\nwant %+v", logData1, expectedLogData1)
	}
	if !reflect.DeepEqual(logData2, expectedLogData2) {
		t.Fatalf("logs2 dont match\nhave %+v\nwant %+v", logData2, expectedLogData2)
	}
}
