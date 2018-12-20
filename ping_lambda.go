package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"net"
	"time"
)

type LambdaEvent struct {
	Name string `json:"name"`
	Env  string `json:"environment"`
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

var svc *cloudwatch.CloudWatch

func HandleRequest(ctx context.Context, ev LambdaEvent) (string, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ev.Ip, ev.Port), 1*time.Second)
	end := time.Now().Sub(start)
	conn.Close()

	var connState = 1.0
	if err != nil {
		connState = 0.0
		fmt.Printf("ERROR: Can't access service %s, endpoint=%s:%s - %s\n", ev.Name, ev.Ip, ev.Port, err.Error())
	}
	r, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("Service/Dial"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("State"),
				Unit:       aws.String("None"),
				Value:      aws.Float64(connState),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("ServiceName"),
						Value: aws.String(ev.Name),
					},
					&cloudwatch.Dimension{
						Name:  aws.String("Environment"),
						Value: aws.String(ev.Env),
					},
				},
			},
			&cloudwatch.MetricDatum{
				MetricName: aws.String("Latency"),
				Unit:       aws.String("Milliseconds"),
				Value:      aws.Float64(float64(end) / float64(time.Millisecond)),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("ServiceName"),
						Value: aws.String(ev.Name),
					},
					&cloudwatch.Dimension{
						Name:  aws.String("Environment"),
						Value: aws.String(ev.Env),
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return "", err
	} else {
		fmt.Println("Res: %s", r)
	}
	return fmt.Sprintf("Done %s!", ev.Name), nil
}

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create new cloudwatch client.
	svc = cloudwatch.New(sess)
}

func main() {
	lambda.Start(HandleRequest)
}
