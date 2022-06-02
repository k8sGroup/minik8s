package object

import (
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"go.uber.org/atomic"
	"strings"
)

type Job2Pod struct {
	PodName string
}

type GPUJob struct {
	Metadata ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     JobSpec    `json:"spec" yaml:"spec"`
}

type JobSpec struct {
	SlurmConfig JobConfig `json:"slurmConfig" yaml:"slurmConfig"`
	Commands    []string  `json:"commands" yaml:"commands"`
	ZipPath     string    `json:"zipPath" yaml:"zipPath"`
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
	if config.NTasks > 0 {
		model = append(model, fmt.Sprintf("#SBATCH -n %d", config.NTasks))
	}
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
	username0 string = "stu625"
	password0        = "%Fi4X^n@"
	username1        = "stu626"
	password1        = "2#Sp8Ejw"
	username2        = "stu627"
	password2        = "YYT!Y9ok"
)

var HostSySet = mapset.NewSet[string]("64c512g", "a100")
var HostPiAndAISet = mapset.NewSet[string]("small", "debug", "cpu", "dgx2")
var HostArmSet = mapset.NewSet[string]("arm128c256g")

type Account struct {
	username       string
	password       string
	host           string
	remoteBasePath string
}

func NewGPUAccount(username string, password string) *Account {
	return &Account{
		username: username,
		password: password,
	}
}

func (account *Account) SetRemoteBasePath(host string) error {
	switch host {
	case HostSy:
		account.host = host
		account.remoteBasePath = fmt.Sprintf("/dssg/home/acct-stu/%s", account.username)
		return nil
	case HostPiAndAI:
		account.host = host
		account.remoteBasePath = fmt.Sprintf("/lustre/home/acct-stu/%s", account.username)
		return nil
	case HostArm:
		account.host = host
		account.remoteBasePath = fmt.Sprintf("/lustre/home/acct-stu/%s", account.username)
		return nil
	default:
		return errors.New("unknown host type")
	}
}

func (account *Account) GetUsername() string {
	return account.username
}

func (account *Account) GetPassword() string {
	return account.password
}

func (account *Account) GetHost() string {
	return account.host
}

func (account *Account) GetRemoteBasePath() string {
	return account.remoteBasePath
}

type JobZipFile struct {
	Key   string `json:"key" yaml:"key"` // format : job-uuid
	Slurm []byte `json:"slurm" yaml:"slurm"`
	Zip   []byte `json:"zip" yaml:"zip"`
}

type AccountAllocator struct {
	counter *atomic.Uint64
}

func NewAccountAllocator() *AccountAllocator {
	return &AccountAllocator{
		counter: atomic.NewUint64(0),
	}
}

func (a *AccountAllocator) Allocate(partition string) (*Account, error) {
	var host string
	if HostSySet.Contains(strings.ToLower(partition)) {
		host = HostSy
	} else if HostPiAndAISet.Contains(strings.ToLower(partition)) {
		host = HostPiAndAI
	} else if HostArmSet.Contains(strings.ToLower(partition)) {
		host = HostArm
	} else {
		return nil, errors.New("illegal partition")
	}
	t := a.counter.Add(1)
	var account *Account
	switch t % 3 {
	case 0:
		account = NewGPUAccount(username0, password0)
	case 1:
		account = NewGPUAccount(username1, password1)
	case 2:
		account = NewGPUAccount(username2, password2)
	default:
		return nil, errors.New("allocation error")
	}
	err := account.SetRemoteBasePath(host)
	if err != nil {
		return nil, err
	}
	return account, nil
}
