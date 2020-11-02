package asset

func InitCache() {
	initUser();
	initEdgeDevice()
	// 5分钟执行一次
	//c := time.Tick(5 * 60 * time.Second)
	//for {
	//	<- c
	//	initUser();
	//	initEdgeDevice()
	//}
}
