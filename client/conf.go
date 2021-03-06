package cbfsclient

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/couchbaselabs/cbfs/config"
	"github.com/dustin/httputil"
)

func (c Client) confURL() string {
	return c.URLFor(".cbfs/config/")
}

// Get the current configuration.
func (c Client) GetConfig() (rv cbfsconfig.CBFSConfig, err error) {
	err = getJsonData(c.confURL(), &rv)
	return
}

// Set a configuration parameter by name.
func (c Client) SetConfigParam(key, val string) error {
	conf, err := c.GetConfig()
	if err != nil {
		return err
	}

	err = conf.SetParameter(key, val)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&conf)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.confURL(),
		bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		return httputil.HTTPError(res)
	}
	return nil
}
