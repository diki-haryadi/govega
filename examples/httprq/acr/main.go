package main

import (
	"bitbucket.org/rctiplus/vegapunk/httprq"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Video struct {
	ID                  int64  `json:"id"`
	ContestantID        int64  `json:"contestant_id"`
	VideoID             string `json:"video_id"`
	VideoTitle          string `json:"video_title"`
	Status              string `json:"status"`
	StatusMute          int64  `json:"status_mute"`
	SourceVideoOriginal string `json:"source_video_original"`
	VideoSource         string `json:"video_source"`
	SourceVideoMuted    string `json:"source_video_muted"`
	SourceVideoFull     string `json:"source_video_full"`
	SourceAudio         string `json:"source_audio"`
	CheckerStatus       string `json:"checker_status"`
	CompetitionID       int    `json:"competition_id"`
}

type ACRRequest struct {
	DataType string `json:"data_type"`
	Url      string `json:"url"`
	Name     string `json:"name"`
}

type DataResponseUploadACR struct {
	Data struct {
		Uid       int         `json:"uid"`
		Cid       int         `json:"cid"`
		Name      string      `json:"name"`
		Duration  int         `json:"duration"`
		Uri       string      `json:"uri"`
		DataType  string      `json:"data_type"`
		Engine    int         `json:"engine"`
		Count     int         `json:"count"`
		State     int         `json:"state"`
		Url       string      `json:"url"`
		UpdatedAt time.Time   `json:"updated_at"`
		CreatedAt time.Time   `json:"created_at"`
		Id        string      `json:"id"`
		Total     int         `json:"total"`
		Results   interface{} `json:"results"`
		Detail    string      `json:"detail"`
	} `json:"data"`
}

type CallbackARC1 struct {
	Cid     int    `json:"cid"`
	FileId  string `json:"file_id"`
	Results struct {
		CoverSongs string `json:"cover_songs"`
	} `json:"results"`
	State int `json:"state"`
}

func main() {
	//https://short.rctiplus.id/vod-e5a2a2/8a67f120769071eebfea97c6360c0102/9a909c2a92dc40a08f7edc5c5613dc15-2f3d2563027e0c99a737e87d4de8886d-sq.mp3https://short.rctiplus.id/vod-e5a2a2/8a67f120769071eebfea97c6360c0102/9a909c2a92dc40a08f7edc5c5613dc15-2f3d2563027e0c99a737e87d4de8886d-sq.mp3
	video := Video{
		ID:                  2510,
		ContestantID:        3179,
		VideoID:             "8d424bba62954e66a2bcd12e7393b04d",
		VideoTitle:          "indah Sapta yunita,31 tahun, kabupaten Pemalang.. #Cidro #the next Didi kempot",
		Status:              "ready",
		StatusMute:          0,
		SourceVideoOriginal: "https://vcdn3.rctiplus.id/customerTrans/d17785282cdea5ccbc5c75a50fc10d24/38ee2cb1-17592788a8f-0013-7f2c-396-a743f.mov",
		VideoSource:         "https://vcdn3.rctiplus.id/8d424bba62954e66a2bcd12e7393b04d/75f3a885b5cc400aa7794bfae3a0618d.m3u8",
		SourceVideoMuted:    "https://vcdn3.rctiplus.id/8d424bba62954e66a2bcd12e7393b04d/75f3a885b5cc400aa7794bfae3a0618d.m3u8",
		SourceVideoFull:     "https://vcdn3.rctiplus.id/8d424bba62954e66a2bcd12e7393b04d/75f3a885b5cc400aa7794bfae3a0618d.m3u8",
		SourceAudio:         "https://short.rctiplus.id/vod-e5a2a2/20dba241d16871ee8b071426d1810102/1bde4be8a7184044b531602455078d07-836c2357442a53bfcd468c74547f6779-sq.mp3",
		CheckerStatus:       "default",
		CompetitionID:       0,
	}
	// update to checking
	// get videos
	// hit post acr and waiting for callback
	//err := c.repositoryVideos.UpdatesStatusConsumerVideos(ctx, &entities.Video{
	//	ID:            video.ID,
	//	Status:        "ready",
	//	CheckerStatus: "checking",
	//})

	//if err != nil {
	//	e := apm.CaptureError(ctx, err)
	//	e.Send()
	//
	//	fmt.Println(err.Error())
	//}
	//
	//excComps, err := c.repositoryVideos.GetExcludeComp(ctx)
	//
	//// Flag to indicate whether the value is found
	var found bool

	// Iterate over the array to find the value
	//for _, value := range excComps {
	//	if value == video.CompetitionID {
	//		found = true
	//		break
	//	}
	//}

	found = true
	//var messageError string
	//var statusCodeError int
	//var body map[string]interface{}
	response := DataResponseUploadACR{}
	//var label string
	//var title string
	//var artist string
	var metadata string
	//var ctx context.Context
	//var labelID int64

	//	curl --location 'https://api-v2.acrcloud.com/api/fs-containers/17150/files' \
	//	--header 'Accept: application/json' \
	//	--header 'Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiI3IiwianRpIjoiMGJiYmRhNWMxY2VhMWNmMGM3YTQ0NDlkN2Q0NjZhYTY4NjUwZjNjZTM4MTU0MWJkNjRlMzlhOTg1YWZkMzdjZmI3MjU3NzMzMjhmMzA5NjkiLCJpYXQiOjE3MDg1ODU4MjMuNzM1NTg1LCJuYmYiOjE3MDg1ODU4MjMuNzM1NTg4LCJleHAiOjIwMjQyMDUwMjMuNjg3OTk2LCJzdWIiOiIxNzIwODIiLCJzY29wZXMiOlsiKiIsIndyaXRlLWFsbCIsInJlYWQtYWxsIiwiYnVja2V0cyIsIndyaXRlLWJ1Y2tldHMiLCJyZWFkLWJ1Y2tldHMiLCJhdWRpb3MiLCJ3cml0ZS1hdWRpb3MiLCJyZWFkLWF1ZGlvcyIsImNoYW5uZWxzIiwid3JpdGUtY2hhbm5lbHMiLCJyZWFkLWNoYW5uZWxzIiwiYmFzZS1wcm9qZWN0cyIsIndyaXRlLWJhc2UtcHJvamVjdHMiLCJyZWFkLWJhc2UtcHJvamVjdHMiLCJ1Y2YiLCJ3cml0ZS11Y2YiLCJyZWFkLXVjZiIsImRlbGV0ZS11Y2YiLCJibS1wcm9qZWN0cyIsImJtLWNzLXByb2plY3RzIiwid3JpdGUtYm0tY3MtcHJvamVjdHMiLCJyZWFkLWJtLWNzLXByb2plY3RzIiwiYm0tYmQtcHJvamVjdHMiLCJ3cml0ZS1ibS1iZC1wcm9qZWN0cyIsInJlYWQtYm0tYmQtcHJvamVjdHMiLCJmaWxlc2Nhbm5pbmciLCJ3cml0ZS1maWxlc2Nhbm5pbmciLCJyZWFkLWZpbGVzY2FubmluZyIsIm1ldGFkYXRhIiwicmVhZC1tZXRhZGF0YSJdfQ.kplGUGz8jkQhu4dxbZTLLvCT-WGcICbpWoUt4QGiQrka6LeM02TBUHpXD42Hp-D6WdCX4I_mLjqnbWROiDN_r0Ee8tfQLUX_jCblOrUyfPLp-s4YYNojm-vfWE6i7Occn87kz-H7RNIqHovMg8p8hzH0jDZuEq30KpkmeN09OVzFbo4pBQfC5hpvDJfCOvBAPJjUi7kILTnXlcMF0do8De-8fFWm59gV3m4XqJ2cgZkvvmJq3mtZPZvDCgf7u3_MQ7ZM7wY7hvPTIguq_hWFWofoBYWd6HVAQUh_0w_x3CHURzbc0gYx2ZOh4HTmSgWz7x3Zyid6ybgRDZyCwCuzc-5u6jw4Wg0F1qJCjFAeNhD3Tjh1JpzG4xHFyq_160fipqBdp7JT9dpjVi4nISmVuTQyBE4d22egh7VM046u7eg_T5Pm2PGCVorzgJ5_qPMtgtMXavR_rpUd1lWSebYeVhgobxWMHejXsCBlAXMMNxjdrttlfxv96XpWE9R4d3x6fzQpzKRxjOoPjp_DhGB3vQrrODHliu3Abyd9LxQ04aQ7Pv_zjQE1tB7RgHT6l5CuhWwQCsyPuNw-WrBiDz_ZZKdrN-lKnw1oGBpGsAnFAuB9XzYwiMXP5FIx3TKBSGtiiyaAFwYPjJ_LsKtBJcCfSUqOM4lHDkiojw27rfNE2KE' \
	//	--header 'Content-Type: application/json' \
	//	--data '{
	//	"data_type": "audio_url",
	//		"url": "https://short.rctiplus.id/vod-e5a2a2/8a67f120769071eebfea97c6360c0102/9a909c2a92dc40a08f7edc5c5613dc15-2f3d2563027e0c99a737e87d4de8886d-sq.mp3",
	//		"name": "ee1"
	//}'

	if found {
		request := ACRRequest{
			DataType: "audio_url",
			Url:      video.SourceAudio,
			Name:     video.VideoID,
		}

		requestBody, _ := json.Marshal(request)

		url := "https://api-v2.acrcloud.com/api/fs-containers/" +
			fmt.Sprint(17210) + "/files"

		err := httprq.Post(url).
			WithContext(context.Background()).
			AddHeader("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiI3IiwianRpIjoiNGRmZWFmODU3Yzc2YTY3MWU4NTc3NjBlODk4NzU1NmQ5MmZjYjMzNTA3MTU5YWUyMWVjNjE0ZWI4OGRlZDQzZTgwYzVjMjJkNTExYWIyNTMiLCJpYXQiOjE3MDkyNjU5MzYuMTU1OTk0LCJuYmYiOjE3MDkyNjU5MzYuMTU1OTk5LCJleHAiOjIwMjQ3OTg3MzMuOTQ1NzMyLCJzdWIiOiIxNzI2MDQiLCJzY29wZXMiOlsiKiIsIndyaXRlLWFsbCIsInJlYWQtYWxsIiwiYnVja2V0cyIsIndyaXRlLWJ1Y2tldHMiLCJyZWFkLWJ1Y2tldHMiLCJhdWRpb3MiLCJ3cml0ZS1hdWRpb3MiLCJyZWFkLWF1ZGlvcyIsImNoYW5uZWxzIiwid3JpdGUtY2hhbm5lbHMiLCJyZWFkLWNoYW5uZWxzIiwiYmFzZS1wcm9qZWN0cyIsIndyaXRlLWJhc2UtcHJvamVjdHMiLCJyZWFkLWJhc2UtcHJvamVjdHMiLCJ1Y2YiLCJ3cml0ZS11Y2YiLCJyZWFkLXVjZiIsImRlbGV0ZS11Y2YiLCJibS1wcm9qZWN0cyIsImJtLWNzLXByb2plY3RzIiwid3JpdGUtYm0tY3MtcHJvamVjdHMiLCJyZWFkLWJtLWNzLXByb2plY3RzIiwiYm0tYmQtcHJvamVjdHMiLCJ3cml0ZS1ibS1iZC1wcm9qZWN0cyIsInJlYWQtYm0tYmQtcHJvamVjdHMiLCJmaWxlc2Nhbm5pbmciLCJ3cml0ZS1maWxlc2Nhbm5pbmciLCJyZWFkLWZpbGVzY2FubmluZyIsIm1ldGFkYXRhIiwicmVhZC1tZXRhZGF0YSJdfQ.VTYIiE__zjBF2CLk24lTJ9J2CyGuHlHCRL0YsqnAEYLgjKNq81-cLM2LT2cIh4Ql1rweYhbcDHcx0gbHss6c5ZAaCsEKpcYVGfn3wIxHDRdZOkFDarb96TCfSh-WfGaQWBrTstw5cMM7UrqQnbK-DG3cORytToDfto7SamGEJZ7RC_0XqkOVegdpTaLHaX2-2No3R8AcrbaKD1JkPgIx7IqVxMPOAbVmk8bTTkPDwvbtqjx3IiQuQVWPvXb1K57qJTt5MjUY8zRWT59OeAmnNStW-sarDuM7YWCRBpa0ZlTC8g8x_nFUuzknW--gMRFnyZrhxZP3doUtCxyQO56lxVaHx6mcPF60uh8HIiDAAuO2MrwhJvc0HtEpHZeOxbsxXg2nHKm52qJXEwuI-EZ1wYel3_AEIMJloYET4yyvF7LVvgeA99HU2Hqg1sbfK1umlElo8S3uAOFfKKWSU7t3ErhxJU8dDHBFND1pNhtuIL15R_FowM6C-fhl77HQQUolu0dQEy4-6WG_x4x-u0NYviuw1Ab4INQ0A2nSN7HAuxywpJjxd_9bdTfi7BDG1s22l2r2coYbBNbxCPxpn-92SZSCUDvN2yIy0UnbjFimlR5zT5goJZsO7WUoEyRGXOyEot6T6owDxCNIbHw8RNkti2-IE1c-G4GPF_iIA7tQuT4").
			AddHeader("Content-Type", "application/json").
			AddHeader("Accept", "application/json").
			WithBody(bytes.NewBuffer(requestBody)).
			WithRetryStrategyWhenTimout(3).
			WithTimeoutHystrix(
				5000,
				100,
				25).
			Execute().
			Consume(&response)

		if err != nil {
			//e := apm.CaptureError(ctx, err)
			//e.Send()
			//
			//_, errAuditTrail := c.repository.AuditTrail(ctx, &entities.ShazamLog{
			//	VideoID:      video.ID,
			//	Message:      err.Error(),
			//	StatusCode:   500,
			//	ResponseBody: "",
			//})
			//if errAuditTrail != nil {
			//	e := apm.CaptureError(ctx, err)
			//	e.Send()
			//}
			//
			//err = c.repositoryVideos.UpdatesStatusConsumerVideos(ctx, &entities.Video{
			//	ID:            video.ID,
			//	Status:        "ready",
			//	CheckerStatus: "submitted",
			//})
			//if err != nil {
			//	e := apm.CaptureError(ctx, err)
			//	e.Send()
			//}

			return
		}
	}

	if response.Data.Name == "" {
		//c.repository.AuditTrail(ctx, &entities.ShazamLog{
		//	VideoID:      video.ID,
		//	Message:      messageError,
		//	StatusCode:   statusCodeError,
		//	ResponseBody: metadata,
		//})
		//
		//err = c.repositoryVideos.UpdatesStatusConsumerVideos(ctx, &entities.Video{
		//	ID:            video.ID,
		//	Status:        "ready",
		//	CheckerStatus: "submitted",
		//})
		fmt.Printf(
			"\n> Body Nil From Get Shazam And restore to Uploaded : %s \n\n",
			metadata,
		)

		return
	}

	//unchecked
	//checked
	//default
	//err = c.repositoryVideos.UpdatesStatusConsumerVideos(ctx, &entities.Video{
	//	ID:            video.ID,
	//	Status:        "ready",
	//	CheckerStatus: "submitted",
	//})
	fmt.Println("update to submitted")

	fmt.Printf(
		"\n> Successfully consume : %s \n\n",
		video,
	)

	return
}
