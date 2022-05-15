package object

type Resource string

const (
	CPU_RESOURCE    Resource = "cpu"
	MEMORY_RESOURCE          = "memory"
)

/*
JSON EXAMPLE
{
    "status": "success",
    "data": {
        "resultType": "vector",
        "result": [
            {
                "metric": {
                    "__name__": "node_monitor",
                    "instance": "192.168.1.156:9070",
                    "job": "node",
                    "node": "node=test",
                    "pod": "pod=",
                    "resource": "cpu"
                },
                "value": [
                    1652587093.695,
                    "0.004550104166132792"
                ]
            }
        ]
    }
}
*/

type PromQueryRes struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	ResultType  string   `json:"resultType"`
	ResultArray []Result `json:"result"`
}

type Result struct {
	Value []interface{} `json:"value"`
}
