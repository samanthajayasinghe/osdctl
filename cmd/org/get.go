package org

import (
	"fmt"
	"log"
	"os"

	"github.com/openshift-online/ocm-cli/pkg/arguments"
	"github.com/openshift-online/ocm-cli/pkg/dump"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	getAPIPath = "/api/accounts_mgmt/v1/accounts"
)

var (
	getCmd = &cobra.Command{
		Use:           "get",
		Short:         "get oraganization",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {

			cmdutil.CheckErr(SearchOrgByUsers(cmd))
		},
	}
	searchUser       string
	isPartMatch      bool   = false
	searchLikeAppend string = "%"
	seachLikePrepend string
)

func init() {
	// define flags
	getCmd.Flags().StringVarP(&searchUser, "user", "u", "", "search organization by user name ")
	getCmd.Flags().BoolVarP(&isPartMatch, "part-match", "", false, "Part matching user name")
	getCmd.MarkFlagRequired("user")

}

func SearchOrgByUsers(cmd *cobra.Command) error {
	response, err := GetOrgs()

	if err != nil {
		// If the response has errored, likely the input was bad, so show usage
		err := cmd.Help()
		if err != nil {
			return err
		}
		return err
	}

	err = dump.Pretty(os.Stdout, response.Bytes())

	if err != nil {
		// If outputing the data errored, there's likely an internal error, so just return the error
		return err
	}
	return nil
}

func GetOrgs() (*sdk.Response, error) {
	// Create OCM client to talk
	ocmClient := utils.CreateConnection()
	defer func() {
		if err := ocmClient.Close(); err != nil {
			fmt.Printf("Cannot close the ocmClient (possible memory leak): %q", err)
		}
	}()

	// Now get the matching orgs
	return sendRequest(CreateGetOrgsRequest(ocmClient))
}

func CreateGetOrgsRequest(ocmClient *sdk.Connection) *sdk.Request {
	// Create and populate the request:
	request := ocmClient.Get()
	err := arguments.ApplyPathArg(request, getAPIPath)

	if err != nil {
		log.Fatalf("Can't parse API path '%s': %v\n", getAPIPath, err)

	}
	if isPartMatch {
		seachLikePrepend = "%"
	}

	formatMessage := fmt.Sprintf(
		`search=username like '%s%s%s'`,
		seachLikePrepend,
		searchUser,
		searchLikeAppend,
	)
	arguments.ApplyParameterFlag(request, []string{formatMessage})

	return request
}
