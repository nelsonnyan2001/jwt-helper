package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoActions struct {
	CognitoClient *cognitoidentityprovider.Client
}

// SignIn signs in a user to Amazon Cognito using a username and password authentication flow.
func (actor CognitoActions) SignIn(clientId, userName, password string) (*types.AuthenticationResultType, error) {
	output, err := actor.CognitoClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       aws.String(clientId),
		AuthParameters: map[string]string{"USERNAME": userName, "PASSWORD": password},
	})
	return output.AuthenticationResult, err

}

func GetDetails(clientId, username, password string) (*types.AuthenticationResultType, error) {
	ca := CognitoActions{CognitoClient: cognitoidentityprovider.New(
		cognitoidentityprovider.Options{
			APIOptions: aws.NewConfig().APIOptions,
			Region:     "us-west-2"})}
	return ca.SignIn(clientId, username, password)
}
