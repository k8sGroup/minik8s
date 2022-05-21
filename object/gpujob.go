package object

import (
	"errors"
	"fmt"
	"strings"
)

type GPUJob struct {
	Metadata ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     JobSpec    `json:"spec" yaml:"spec"`
}

type JobSpec struct {
	SlurmConfig JobConfig `json:"slurmConfig" yaml:"slurmConfig"`
	Commands    []string  `json:"commands" yaml:"commands"`
}

type JobConfig struct {
	JobName         string `json:"jobName" yaml:"jobName"`
	Partition       string `json:"partition" yaml:"partition"`
	CpusPerTask     int32  `json:"cpusPerTask" yaml:"cpusPerTask"`
	Nodes           int32  `json:"nodes" yaml:"nodes"`
	NTasks          int32  `json:"nTasks" yaml:"nTasks"`
	NTasksPerNode   int32  `json:"nTasksPerNode" yaml:"nTasksPerNode"`
	GenericResource string `json:"gres" yaml:"gres"`
	Output          string `json:"output" yaml:"output"`
	Error           string `json:"error" yaml:"error"`
	Time            string `json:"time" yaml:"time"`
	Array           string `json:"array" yaml:"array"`
	Depend          string `json:"depend" yaml:"depend"`
	MailType        string `json:"mailType" yaml:"mailType"`
	MailUser        string `json:"mailUser" yaml:"mailUser"`
}

func (j *GPUJob) GenerateSlurmScript() []byte {
	var model []string
	config := &j.Spec.SlurmConfig
	model = append(model, "#!/bin/bash")
	model = append(model, fmt.Sprintf("#SBATCH --job-name=%s", config.JobName))
	model = append(model, fmt.Sprintf("#SBATCH --partition=%s", config.Partition))
	if config.CpusPerTask > 0 {
		model = append(model, fmt.Sprintf("#SBATCH --cpus-per-task=%d", config.CpusPerTask))
	}
	if config.Nodes > 0 {
		model = append(model, fmt.Sprintf("#SBATCH --nodes=%d", config.Nodes))
	}
	model = append(model, fmt.Sprintf("#SBATCH -n %d", config.NTasks))
	if config.NTasksPerNode > 0 {
		model = append(model, fmt.Sprintf("#SBATCH --ntasks-per-node=%d", config.NTasksPerNode))
	}
	model = append(model, fmt.Sprintf("#SBATCH --gres=%s", config.GenericResource))
	if config.Output != "" {
		model = append(model, fmt.Sprintf("#SBATCH --output=%s", config.Output))
	}
	if config.Error != "" {
		model = append(model, fmt.Sprintf("#SBATCH --error=%s", config.Error))
	}
	if config.Time != "" {
		model = append(model, fmt.Sprintf("#SBATCH --time=%s", config.Time))
	}
	if config.Array != "" {
		model = append(model, fmt.Sprintf("#SBATCH --array=%s", config.Array))
	}
	if config.Depend != "" {
		model = append(model, fmt.Sprintf("#SBATCH --depend=%s", config.Depend))
	}
	if config.MailType != "" {
		model = append(model, fmt.Sprintf("#SBATCH --mail-type=%s", config.MailType))
	}
	if config.MailUser != "" {
		model = append(model, fmt.Sprintf("#SBATCH --mail-user=%s", config.MailUser))
	}

	for _, cmd := range j.Spec.Commands {
		model = append(model, cmd)
	}

	return []byte(strings.Join(model, "\n"))
}

type JobStatus struct {
	JID    string `json:"jid" yaml:"jid"`
	Status string `json:"status" yaml:"status"`
}

const (
	HostSy      string = "sylogin.hpc.sjtu.edu.cn"
	HostPiAndAI        = "login.hpc.sjtu.edu.cn"
	HostArm            = "armlogin.hpc.sjtu.edu.cn"
)

const (
	Username1 string = "stu625"
	Password1        = "%Fi4X^n@"
	Username2        = "stu626"
	Password2        = "2#Sp8Ejw"
	Username3        = "stu627"
	Password3        = "YYT!Y9ok"
)

type Account struct {
	username string
	password string
}

func NewGPUAccount(username string, password string) *Account {
	return &Account{
		username: username,
		password: password,
	}
}

func (account *Account) RemoteBasePath(cluster string) (string, error) {
	switch cluster {
	case HostSy:
		return fmt.Sprintf("/dssg/home/acct-stu/%s", account.username), nil
	case HostPiAndAI:
		return fmt.Sprintf("/lustre/home/acct-stu/%s", account.username), nil
	case HostArm:
		return fmt.Sprintf("/lustre/home/acct-stu/%s", account.username), nil
	default:
		return "", errors.New("unknown cluster type")
	}
}

func (account *Account) GetUsername() string {
	return account.username
}

func (account *Account) GetPassword() string {
	return account.password
}
