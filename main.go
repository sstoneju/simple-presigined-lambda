package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type APIGatewayProxyRequest struct {
    Version               string            `json:"version"`
    RouteKey              string            `json:"routeKey"`
    RawPath               string            `json:"rawPath"`
    RawQueryString        string            `json:"rawQueryString"`
    Cookies               []string          `json:"cookies"`
    Headers               map[string]string `json:"headers"`
    QueryStringParameters map[string]string `json:"queryStringParameters"`
    RequestContext        RequestContext    `json:"requestContext"`
    Body                  string            `json:"body"`
    PathParameters        map[string]string `json:"pathParameters"`
    IsBase64Encoded       bool              `json:"isBase64Encoded"`
    StageVariables        map[string]string `json:"stageVariables"`
}

type RequestContext struct {
    AccountId     string          `json:"accountId"`
    ApiId         string          `json:"apiId"`
    Authentication Authentication `json:"authentication"`
    Authorizer     Authorizer     `json:"authorizer"`
    DomainName     string         `json:"domainName"`
    DomainPrefix   string         `json:"domainPrefix"`
    Http           Http           `json:"http"`
    RequestId      string         `json:"requestId"`
    RouteKey       string         `json:"routeKey"`
    Stage          string         `json:"stage"`
    Time           string         `json:"time"`
    TimeEpoch      int64          `json:"timeEpoch"`
}

type Authentication struct {
    ClientCert ClientCert `json:"clientCert"`
}

type Authorizer struct {
    Jwt Jwt `json:"jwt"`
}

type ClientCert struct {
    ClientCertPem string    `json:"clientCertPem"`
    SubjectDN     string    `json:"subjectDN"`
    IssuerDN      string    `json:"issuerDN"`
    SerialNumber  string    `json:"serialNumber"`
    Validity      Validity  `json:"validity"`
}

type Jwt struct {
    Claims map[string]string `json:"claims"`
    Scopes []string          `json:"scopes"`
}

type Http struct {
    Method    string `json:"method"`
    Path      string `json:"path"`
    Protocol  string `json:"protocol"`
    SourceIp  string `json:"sourceIp"`
    UserAgent string `json:"userAgent"`
}

type Validity struct {
    NotBefore string `json:"notBefore"`
    NotAfter  string `json:"notAfter"`
}

// APIGatewayProxyResponse는 API Gateway로부터의 HTTP 응답을 나타냅니다.
type APIGatewayProxyResponse struct {
    StatusCode int               `json:"statusCode"`
    Headers    map[string]string `json:"headers"`
    Body       string            `json:"body,omitempty"`
}

// 접근 가능한 버킷을 정의합니다.
type admitBucket struct {
	BucketName []string	`json:"buketName"`
}

func CheckBucket(bucket string) (bool) {
	admit := admitBucket{
		BucketName: []string{"perfitt-ai-image-dev"},
	}

	for _, v := range admit.BucketName {
		if bucket == v {
			return true
		}
	}
	return false
}

func SignedURL(bucket, key string) (string) {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2")},
	)

	// Create S3 service client
	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(10 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	log.Println("The URL is", urlStr)
	return urlStr
}

func HandleRequest(ctx context.Context, event *APIGatewayProxyRequest) (APIGatewayProxyResponse, error) {
	fmt.Println("ctx: ", ctx)
	fmt.Println("QueryStringParameters: ", event.QueryStringParameters)

	queryParams := event.QueryStringParameters
	apiKey, exists := queryParams["apiKey"]
	if !exists {
		fmt.Println("apiKey is required")
		return APIGatewayProxyResponse{ StatusCode: 500, Body: "apiKey is required"}, nil
	}

	if apiKey != "eyJoZWxsbyI6IndvcmxkIn0K" { // echo '{"hello":"world"}' | base64
		fmt.Println("permission denied")
		return APIGatewayProxyResponse{ StatusCode: 500, Body: "permission denied"}, nil
	}

	bucket, exists := queryParams["bucket"]
	if !exists {
		fmt.Println("bucket is required")
		return APIGatewayProxyResponse{ StatusCode: 500, Body:"bucket is required"}, nil
	}
	if !CheckBucket(bucket) {
		fmt.Println("forbidden bucket")
		return APIGatewayProxyResponse{ StatusCode: 500, Body:"forbidden bucket"}, nil
	}

	key, exists := queryParams["key"]
	if !exists {
		fmt.Println("key is required")
		return APIGatewayProxyResponse{ StatusCode: 500, Body: "key is required"}, nil
	}

	signedUrl := SignedURL(bucket, key)

	response := APIGatewayProxyResponse{
		StatusCode: 200,
		Body: signedUrl,
    }

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}