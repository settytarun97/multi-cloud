// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/api/pkg/s3/datastore"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	"github.com/opensds/multi-cloud/s3/proto"
)

//ObjectGet -
func (s *APIService) ObjectGet(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")
	rangestr := request.HeaderParameter("Range")
	log.Infof("Received request for object get, bucket: %s, object: %s, range: %s\n",
		bucketName, objectKey, rangestr)

	start := 0
	end := 0
	if rangestr != "" {
		index := strings.Index(rangestr, "-")
		startstr := string([]rune(rangestr)[6:index])
		endstr := string([]rune(rangestr)[index+1:])
		start, _ = strconv.Atoi(startstr)
		end, _ = strconv.Atoi(endstr)
	}

	md := map[string]string{common.REST_KEY_OPERATION: common.REST_VAL_DOWNLOAD}
	ctx := common.InitCtxWithVal(request, md)
	object := s3.Object{}
	objectInput := s3.GetObjectInput{Bucket: bucketName, Key: objectKey}
	log.Infof("enter the s3Client download method")
	objectMD, _ := s.s3Client.GetObject(ctx, &objectInput)
	log.Infof("out the s3Client download method")
	var backendname string
	if objectMD != nil {
		object.Size = objectMD.Size
		backendname = objectMD.Backend
	} else {
		log.Errorf("No such object")
		response.WriteError(http.StatusInternalServerError, NoSuchObject.Error())
		return
	}

	object.ObjectKey = objectKey
	object.BucketName = bucketName
	var client datastore.DataStoreAdapter
	if backendname != "" {
		client = getBackendByName(ctx, s, backendname)
	} else {
		client = getBackendClient(ctx, s, bucketName)
	}
	if client == nil {
		response.WriteError(http.StatusInternalServerError, NoSuchBackend.Error())
		return
	}
	log.Infof("enter the download method")
	body, s3err := client.GET(&object, ctx, int64(start), int64(end))
	log.Infof("out  the download method")
	if s3err != NoError {
		response.WriteError(http.StatusInternalServerError, s3err.Error())
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	response.Write(buf.Bytes())
}
