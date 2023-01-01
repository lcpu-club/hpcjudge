package main

import (
	"fmt"
	"strings"
)

func main() {
	// m, _ := minio.New("127.0.0.1:9000", &minio.Options{
	// 	Creds: credentials.NewStaticV4("hpc", "hpc@devel", ""),
	// })
	// for n := range m.ListenBucketNotification(context.Background(), "solutions", "", "result.json", []string{
	// 	"s3:ObjectCreated:*",
	// }) {
	// 	fmt.Printf("%#v\r\n%#v\r\n", n.Err, n.Records[0].S3)
	// 	fmt.Println(n.Records[0].S3.Object.Key, n.Records[0].S3.Object.VersionID)
	// }
	fmt.Println(strings.Cut("a/bb/c", "/"))
}
