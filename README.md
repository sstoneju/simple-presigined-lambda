# simple-presigined-lambda
간단하게 presigned_url을 사용할 수 있게 도와주는 lambda 코드

# Lambda에 배포하기
```sh
# lambda x64로 배포하기
$ GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go
$ zip myFunction.zip bootstrap
$ aws lambda update-function-code --function-name simple-presigned-url-s3 --zip-file fileb://myFunction.zip
```
