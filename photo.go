package vkapi

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

type PhotoAttachment struct {
	ID        int    `json:"id"`
	AID       int    `json:"album_id"`
	OwnerID   int    `json:"owner_id"`
	Photo75   string `json:"photo_75"`
	Photo130  string `json:"photo_130"`
	Photo604  string `json:"photo_604"`
	Photo807  string `json:"photo_807"`
	Photo1280 string `json:"photo_1280"`
	Photo2560 string `json:"photo_2560"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Text      string `json:"text"`
	Created   int64  `json:"created"`
	AccessKey string `json:"access_key"`
}

type photoUploadServer struct {
	UploadURL string `json:"upload_url"`
	AlbumID   int    `json:"album_id"`
	UserID    int    `json:"user_id"`
}

type photoWallUploadResult struct {
	Server int             `json:"server"`
	Hash   string          `json:"hash"`
	Photo  json.RawMessage `json:"photo"`
}

func (client *VKClient) photoGetUploadServer(params url.Values, method string) (*photoUploadServer, error) {
	resp, err := client.MakeRequest(method, params)
	if err != nil {
		return nil, err
	}

	data := new(photoUploadServer)
	json.Unmarshal(resp.Response, data)

	return data, nil
}

func (client *VKClient) photoUpload(params url.Values, files []string, method string) (*photoWallUploadResult, error) {
	serverInfo, err := client.photoGetUploadServer(params, method)
	if err != nil {
		return nil, err
	}

	req, err := client.getPhotoMultipartReq(serverInfo.UploadURL, files)
	if err != nil {
		return nil, err
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	uploadData := new(photoWallUploadResult)
	json.Unmarshal(body, uploadData)
	escaped := strings.Replace(string(uploadData.Photo), "\\", "", -1)
	escaped = escaped[1 : len(escaped)-1]
	uploadData.Photo = []byte(escaped)

	return uploadData, nil
}

func (client *VKClient) photoWallUpload(groupID string, files []string) (*photoWallUploadResult, error) {
	params := url.Values{}
	params.Set("group_id", groupID)

	return client.photoUpload(params, files, "photos.getWallUploadServer")
}

func (client *VKClient) photoMessagesUpload(peerID string, files []string) (*photoWallUploadResult, error) {
	params := url.Values{}
	params.Set("peer_id", peerID)

	return client.photoUpload(params, files, "photos.getMessagesUploadServer")
}

func (client *VKClient) UploadGroupWallPhotos(groupID int, files []string) ([]*PhotoAttachment, error) {
	gidStr := strconv.Itoa(groupID)
	if gidStr[0] == '-' {
		gidStr = gidStr[1:]
	}
	uploadData, err := client.photoWallUpload(gidStr, files)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("group_id", gidStr)
	params.Set("photo", string(uploadData.Photo))
	params.Set("server", strconv.Itoa(uploadData.Server))
	params.Set("hash", uploadData.Hash)

	resp, err := client.MakeRequest("photos.saveWallPhoto", params)
	if err != nil {
		return nil, err
	}

	var photos []*PhotoAttachment
	json.Unmarshal(resp.Response, &photos)

	return photos, err
}

func (client *VKClient) UploadMessagesPhotos(peerID int, files []string) ([]*PhotoAttachment, error) {
	pidStr := strconv.Itoa(peerID)
	if pidStr[0] == '-' {
		pidStr = pidStr[1:]
	}
	uploadData, err := client.photoMessagesUpload(pidStr, files)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("photo", string(uploadData.Photo))
	params.Set("server", strconv.Itoa(uploadData.Server))
	params.Set("hash", uploadData.Hash)

	resp, err := client.MakeRequest("photos.saveMessagesPhoto", params)
	if err != nil {
		return nil, err
	}

	var photos []*PhotoAttachment
	json.Unmarshal(resp.Response, &photos)

	return photos, err
}

func (client *VKClient) GetPhotosString(photos []*PhotoAttachment) string {
	s := []string{}

	for _, p := range photos {
		s = append(s, "photo"+strconv.Itoa(p.OwnerID)+"_"+strconv.Itoa(p.ID))
	}

	return strings.Join(s, ",")
}
