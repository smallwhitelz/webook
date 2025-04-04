package test

import "encoding/json"

type EventCheck interface {
	Check(want string, actual string) (any, any)
}

type ArticleEvent struct {
	Uid   string `json:"uid"`
	Aid   string `json:"aid"`
	Title string `json:"title"`
}

func (a ArticleEvent) Check(want string, actual string) (any, any) {
	wantVal := ArticleEvent{}
	json.Unmarshal([]byte(want), &wantVal)
	actualVal := ArticleEvent{}
	json.Unmarshal([]byte(actual), &actualVal)
	return wantVal, actualVal
}

type LikeEvent struct {
	Liked string `json:"liked"`
	Liker string `json:"liker"`
	BizID string `json:"bizId"`
	Biz   string `json:"biz"`
}

func (l LikeEvent) Check(want string, actual string) (any, any) {
	wantVal := LikeEvent{}
	json.Unmarshal([]byte(want), &wantVal)
	actualVal := LikeEvent{}
	json.Unmarshal([]byte(actual), &actualVal)
	return wantVal, actualVal
}

type FollowEvent struct {
	Followee string `json:"followee"`
	Follower string `json:"follower"`
}

func (f FollowEvent) Check(want string, actual string) (any, any) {
	wantVal := FollowEvent{}
	json.Unmarshal([]byte(want), &wantVal)
	actualVal := FollowEvent{}
	json.Unmarshal([]byte(actual), &actualVal)
	return wantVal, actualVal
}
