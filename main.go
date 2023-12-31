/*
 * © 2023 Khulnasoft Limited
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	capture "github.com/khulnasoft-lab/vulnmap-iac-capture/cmd/vulnmap-iac-capture"
)

var (
	version string
	commit  string
)

func main() {
	flag.BoolP("debug", "d", false, "Show debug information")
	//flag.Bool("http-tls-skip-verify", false, "If set, skip client validation of TLS certificates") Disabled for the closed beta
	flag.String("api-rest-url", "https://api.vulnmap.khulnasoft.com", "Url for Vulnmap REST API")
	flag.String("path", ".", "Path to look for Terraform state files (can be a file, a directory or a glob pattern)")
	flag.String("api-rest-token", "", "Auth token for the API Usage (Required)")
	flag.String("org", "", "Organization public id (Required)")
	flag.Bool("stdin", false, "Read Terraform state from the standard input instead of path")
	flag.Parse()

	// normalize flag with - in the name to make it easier to match with env
	f := flag.CommandLine
	normalizeFunc := f.GetNormalizeFunc()
	f.SetNormalizeFunc(func(fs *flag.FlagSet, name string) pflag.NormalizedName {
		result := normalizeFunc(fs, name)
		name = strings.ReplaceAll(string(result), "-", "_")
		return pflag.NormalizedName(name)
	})

	viper.SetEnvPrefix("VULNMAP_IAC_CAPTURE")
	must(viper.BindPFlags(flag.CommandLine))
	must(viper.BindEnv("debug"))
	must(viper.BindEnv("org"))
	must(viper.BindEnv("path"))
	must(viper.BindEnv("http_tls_skip_verify"))
	must(viper.BindEnv("api_rest_url"))
	must(viper.BindEnv("api_rest_token"))
	must(viper.BindEnv("stdin"))

	logger := log.New(os.Stderr, "", 0)

	debug := viper.GetBool("debug")
	if !debug {
		logger.SetOutput(io.Discard)
	}

	logger.Printf("vulnmap-iac-capture %s (%s)\n", version, commit)

	command := capture.Command{
		Logger:            logger,
		Org:               viper.GetString("org"),
		StatePath:         viper.GetString("path"),
		HTTPTLSSkipVerify: viper.GetBool("http_tls_skip_verify"),
		APIURL:            viper.GetString("api_rest_url"),
		APIToken:          viper.GetString("api_rest_token"),
		ExtraSSlCerts:     os.Getenv("NODE_EXTRA_CA_CERTS"), // we still want to read this one without prefix to work with vulnmap node cli integration
		StateFromStdin:    viper.GetBool("stdin"),
	}

	os.Exit(command.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
