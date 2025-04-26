// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

package config

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const EnvPrefix = "CAM"

// DefaultCollectionWorkloadID contains the default UUID used for the workload collection module when using the
// auto-create feature.
const DefaultCollectionWorkloadID = "f5ecde04-1ab1-47c8-ad93-6ac241b3e72e"

// DefaultCollectionCommsecID contains the default UUID used for the commsec collection module when using the
// auto-create feature.
const DefaultCollectionCommsecID = "4deec3fd-43f0-40ec-b75e-dab0e7528e09"

// DefaultCollectionIntegrityID contains the default UUID used for the integrity collection module when using the
// auto-create feature.
const DefaultCollectionIntegrityID = "6e58e7d6-774f-4b19-a8e0-a42432023f5c"

// DefaultCollectionAuthSecID contains the default UUID used for the auth security collection module when using the
// auto-create feature.
const DefaultCollectionAuthSecID = "56dd78b5-33de-462e-9b26-8f6e801079e7"

// InitConfig initializes the viper config with sensible defaults for all of the
// CAM modules. It enables loading configuration settings from a config file
// named cam.yaml as well as the environment variables prefixed with EnvPrefix.
func InitConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix(EnvPrefix)
	viper.SetConfigName("cam")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

// AddFlagUint16 adds an uint16 flag to the command and registers with viper and cobra.
func AddFlagUint16(cmd *cobra.Command, flag string, def uint16, usage string) (u *uint16) {
	u = cmd.Flags().Uint16(flag, def, usage)

	_ = viper.BindPFlag(flag, cmd.Flags().Lookup(flag))

	return
}

// AddFlagString adds an uint16 flag to the command and registers with viper and cobra.
func AddFlagString(cmd *cobra.Command, flag string, def string, usage string) (s *string) {
	s = cmd.Flags().String(flag, def, usage)

	_ = viper.BindPFlag(flag, cmd.Flags().Lookup(flag))

	return
}

// AddFlagStringSlice adds an uint16 flag to the command and registers with viper and cobra.
func AddFlagStringSlice(cmd *cobra.Command, flag string, def []string, usage string) (s *[]string) {
	s = cmd.Flags().StringSlice(flag, def, usage)

	_ = viper.BindPFlag(flag, cmd.Flags().Lookup(flag))

	return
}

// AddFlagBool adds an uint16 flag to the command and registers with viper and cobra.
func AddFlagBool(cmd *cobra.Command, flag string, def bool, usage string) (b *bool) {
	b = cmd.Flags().Bool(flag, def, usage)

	_ = viper.BindPFlag(flag, cmd.Flags().Lookup(flag))

	return
}
