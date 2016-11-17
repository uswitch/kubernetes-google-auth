package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type User struct {
	Token           string `yaml:"token,omitempty"`
	CertificateData string `yaml:"client-certificate-data,omitempty"`
	KeyData         string `yaml:"client-key-data,omitempty"`
	Password        string `yaml:"password,omitempty"`
	Username        string `yaml:"username,omitempty"`
}

type UserSpec struct {
	Name string `yaml:"name"`
	User *User  `yaml:"user"`
}

type KubeConfig struct {
	APIVersion     string                      `yaml:"apiVersion"`
	Kind           string                      `yaml:"kind"`
	CurrentContext string                      `yaml:"current-context,omitempty"`
	Users          []*UserSpec                 `yaml:"users"`
	Clusters       []*ClusterSpec              `yaml:"clusters"`
	Contexts       []*ContextSpec              `yaml:"contexts"`
	Preferences    map[interface{}]interface{} `yaml:"preferences"`

	filepath string
}

type Cluster struct {
	CertificateAuthority string `yaml:"certificate-authority-data,omitempty"`
	Server               string `yaml:"server"`
}

type ClusterSpec struct {
	Name    string   `yaml:"name"`
	Cluster *Cluster `yaml:"cluster"`
}

type Context struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type ContextSpec struct {
	Name    string   `yaml:"name"`
	Context *Context `yaml:"context"`
}

func ReadKubeConfig(path string) (*KubeConfig, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config KubeConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	config.filepath = path

	return &config, nil
}

func (c *KubeConfig) findUser(email string) *UserSpec {
	for _, spec := range c.Users {
		if spec.Name == email {
			return spec
		}
	}
	return nil
}

func newUserSpec(email, token string) *UserSpec {
	return &UserSpec{
		Name: email,
		User: &User{
			Token: token,
		},
	}
}

func (c *KubeConfig) UpdateUser(email, token string) {
	spec := c.findUser(email)
	if spec != nil {
		spec.User.Token = token
	} else {
		c.Users = append(c.Users, newUserSpec(email, token))
	}
}

func (c *KubeConfig) findContext(cluster string) *ContextSpec {
	for _, spec := range c.Contexts {
		if spec.Context.Cluster == cluster {
			return spec
		}
	}
	return nil
}

func (c *KubeConfig) UpdateContextUser(user, contextName string) *ContextSpec {
	ctx := c.findContext(contextName)
	if ctx != nil {
		ctx.Context.User = user
		return ctx
	}

	context := &Context{Cluster: contextName, User: user}
	spec := &ContextSpec{Name: contextName, Context: context}
	c.Contexts = append(c.Contexts, spec)
	return spec
}

func (c *KubeConfig) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.filepath, bytes, 0)
}
