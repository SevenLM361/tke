/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package authzwebhook

import (
	"bytes"
	"io/ioutil"

	"github.com/pkg/errors"
	"tkestack.io/tke/pkg/platform/provider/baremetal/constants"
	"tkestack.io/tke/pkg/util/ssh"
	"tkestack.io/tke/pkg/util/template"

	installerconstants "tkestack.io/tke/cmd/tke-installer/app/installer/constants"
)

const (
	authzWebhookConfig = `
apiVersion: v1
kind: Config
clusters:
  - name: tke
    cluster:
      server: {{.AuthzEndpoint}}
      insecure-skip-tls-verify: true
users:
  - name: admin-cert
    user:
      client-certificate: {{.WebhookCertFile}}
      client-key: {{.WebhookKeyFile}}
current-context: tke
contexts:
- context:
    cluster: tke
    user: admin-cert
  name: tke
`
)

type Option struct {
	AuthzWebhookEndpoint string
	IsGlobalCluster      bool
}

func Install(s ssh.Interface, option *Option) error {
	authzWebhookConfig, err := template.ParseString(authzWebhookConfig, map[string]interface{}{
		"AuthzEndpoint": option.AuthzWebhookEndpoint,
		"WebhookCertFile": constants.WebhookCertFile,
		"WebhookKeyFile":  constants.WebhookKeyFile,
	})
	if err != nil {
		return errors.Wrap(err, "parse authzWebhookConfig error")
	}

	err = s.WriteFile(bytes.NewReader(authzWebhookConfig), constants.KubernetesAuthzWebhookConfigFile)
	if err != nil {
		return err
	}
	basePath := constants.AppCertDir
	if option.IsGlobalCluster {
		basePath = installerconstants.DataDir
	}
	webhookCertData, err := ioutil.ReadFile(basePath + constants.WebhookCertName)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(webhookCertData), constants.WebhookCertFile)
	if err != nil {
		return err
	}
	webhookKeyData, err := ioutil.ReadFile(basePath + constants.WebhookKeyName)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(webhookKeyData), constants.WebhookKeyFile)
	if err != nil {
		return err
	}

	return nil
}
