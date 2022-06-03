package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	keyArn := flag.Arg(0)
	if keyArn == "" {
		log.Fatalf("Missing key ARN")
	}

	src := []byte(fmt.Sprintf("aws-provider-token-%s", randSeq(20)))
	buf := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(buf, src)

	// Create an AWS KMS service client
	cli := kms.NewFromConfig(cfg)
	ctx := context.Background()

	so, err := cli.Sign(ctx, &kms.SignInput{
		KeyId:            aws.String(keyArn),
		Message:          buf,
		SigningAlgorithm: types.SigningAlgorithmSpecRsassaPssSha256,
	})
	if err != nil {
		log.Fatal(err)
	}

	signature := base64.RawURLEncoding.EncodeToString(so.Signature)
	token := strings.Join([]string{string(buf), signature}, ".")
	log.Printf("Signed JWT: %s", token)
}
