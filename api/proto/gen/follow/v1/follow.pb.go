// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: follow/v1/follow.proto

package followv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetFollowerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Followee      int64                  `protobuf:"varint,1,opt,name=followee,proto3" json:"followee,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFollowerRequest) Reset() {
	*x = GetFollowerRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFollowerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFollowerRequest) ProtoMessage() {}

func (x *GetFollowerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFollowerRequest.ProtoReflect.Descriptor instead.
func (*GetFollowerRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{0}
}

func (x *GetFollowerRequest) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

type GetFollowerResponse struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	FollowRelations []*FollowRelation      `protobuf:"bytes,1,rep,name=follow_relations,json=followRelations,proto3" json:"follow_relations,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *GetFollowerResponse) Reset() {
	*x = GetFollowerResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFollowerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFollowerResponse) ProtoMessage() {}

func (x *GetFollowerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFollowerResponse.ProtoReflect.Descriptor instead.
func (*GetFollowerResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{1}
}

func (x *GetFollowerResponse) GetFollowRelations() []*FollowRelation {
	if x != nil {
		return x.FollowRelations
	}
	return nil
}

type FollowStatic struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 被多少人关注
	Followers int64 `protobuf:"varint,1,opt,name=followers,proto3" json:"followers,omitempty"`
	// 自己关注了多少人
	Followees     int64 `protobuf:"varint,2,opt,name=followees,proto3" json:"followees,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowStatic) Reset() {
	*x = FollowStatic{}
	mi := &file_follow_v1_follow_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowStatic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowStatic) ProtoMessage() {}

func (x *FollowStatic) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowStatic.ProtoReflect.Descriptor instead.
func (*FollowStatic) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{2}
}

func (x *FollowStatic) GetFollowers() int64 {
	if x != nil {
		return x.Followers
	}
	return 0
}

func (x *FollowStatic) GetFollowees() int64 {
	if x != nil {
		return x.Followees
	}
	return 0
}

type GetFollowStaticRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Followee      int64                  `protobuf:"varint,1,opt,name=followee,proto3" json:"followee,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFollowStaticRequest) Reset() {
	*x = GetFollowStaticRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFollowStaticRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFollowStaticRequest) ProtoMessage() {}

func (x *GetFollowStaticRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFollowStaticRequest.ProtoReflect.Descriptor instead.
func (*GetFollowStaticRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{3}
}

func (x *GetFollowStaticRequest) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

type GetFollowStaticResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	FollowStatic  *FollowStatic          `protobuf:"bytes,1,opt,name=followStatic,proto3" json:"followStatic,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFollowStaticResponse) Reset() {
	*x = GetFollowStaticResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFollowStaticResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFollowStaticResponse) ProtoMessage() {}

func (x *GetFollowStaticResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFollowStaticResponse.ProtoReflect.Descriptor instead.
func (*GetFollowStaticResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{4}
}

func (x *GetFollowStaticResponse) GetFollowStatic() *FollowStatic {
	if x != nil {
		return x.FollowStatic
	}
	return nil
}

type FollowRelation struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Followee      int64                  `protobuf:"varint,2,opt,name=followee,proto3" json:"followee,omitempty"`
	Follower      int64                  `protobuf:"varint,3,opt,name=follower,proto3" json:"follower,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowRelation) Reset() {
	*x = FollowRelation{}
	mi := &file_follow_v1_follow_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowRelation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowRelation) ProtoMessage() {}

func (x *FollowRelation) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowRelation.ProtoReflect.Descriptor instead.
func (*FollowRelation) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{5}
}

func (x *FollowRelation) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *FollowRelation) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

func (x *FollowRelation) GetFollower() int64 {
	if x != nil {
		return x.Follower
	}
	return 0
}

type GetFolloweeRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 关注者，也就是某人查看自己的关注列表
	Follower      int64 `protobuf:"varint,1,opt,name=follower,proto3" json:"follower,omitempty"`
	Offset        int64 `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit         int64 `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFolloweeRequest) Reset() {
	*x = GetFolloweeRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFolloweeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFolloweeRequest) ProtoMessage() {}

func (x *GetFolloweeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFolloweeRequest.ProtoReflect.Descriptor instead.
func (*GetFolloweeRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{6}
}

func (x *GetFolloweeRequest) GetFollower() int64 {
	if x != nil {
		return x.Follower
	}
	return 0
}

func (x *GetFolloweeRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *GetFolloweeRequest) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type GetFolloweeResponse struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	FollowRelations []*FollowRelation      `protobuf:"bytes,1,rep,name=follow_relations,json=followRelations,proto3" json:"follow_relations,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *GetFolloweeResponse) Reset() {
	*x = GetFolloweeResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFolloweeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFolloweeResponse) ProtoMessage() {}

func (x *GetFolloweeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFolloweeResponse.ProtoReflect.Descriptor instead.
func (*GetFolloweeResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{7}
}

func (x *GetFolloweeResponse) GetFollowRelations() []*FollowRelation {
	if x != nil {
		return x.FollowRelations
	}
	return nil
}

type FollowInfoRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 关注者
	Follower int64 `protobuf:"varint,1,opt,name=follower,proto3" json:"follower,omitempty"`
	// 被关注者
	Followee      int64 `protobuf:"varint,2,opt,name=followee,proto3" json:"followee,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowInfoRequest) Reset() {
	*x = FollowInfoRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowInfoRequest) ProtoMessage() {}

func (x *FollowInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowInfoRequest.ProtoReflect.Descriptor instead.
func (*FollowInfoRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{8}
}

func (x *FollowInfoRequest) GetFollower() int64 {
	if x != nil {
		return x.Follower
	}
	return 0
}

func (x *FollowInfoRequest) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

type FollowInfoResponse struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	FollowRelation *FollowRelation        `protobuf:"bytes,1,opt,name=follow_relation,json=followRelation,proto3" json:"follow_relation,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *FollowInfoResponse) Reset() {
	*x = FollowInfoResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowInfoResponse) ProtoMessage() {}

func (x *FollowInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowInfoResponse.ProtoReflect.Descriptor instead.
func (*FollowInfoResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{9}
}

func (x *FollowInfoResponse) GetFollowRelation() *FollowRelation {
	if x != nil {
		return x.FollowRelation
	}
	return nil
}

type FollowRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 被关注者
	Followee int64 `protobuf:"varint,1,opt,name=followee,proto3" json:"followee,omitempty"`
	// 关注者
	Follower      int64 `protobuf:"varint,2,opt,name=follower,proto3" json:"follower,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowRequest) Reset() {
	*x = FollowRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowRequest) ProtoMessage() {}

func (x *FollowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowRequest.ProtoReflect.Descriptor instead.
func (*FollowRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{10}
}

func (x *FollowRequest) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

func (x *FollowRequest) GetFollower() int64 {
	if x != nil {
		return x.Follower
	}
	return 0
}

type FollowResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowResponse) Reset() {
	*x = FollowResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowResponse) ProtoMessage() {}

func (x *FollowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowResponse.ProtoReflect.Descriptor instead.
func (*FollowResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{11}
}

type CancelFollowRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 被关注者
	Followee int64 `protobuf:"varint,1,opt,name=followee,proto3" json:"followee,omitempty"`
	// 关注者
	Follower      int64 `protobuf:"varint,2,opt,name=follower,proto3" json:"follower,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelFollowRequest) Reset() {
	*x = CancelFollowRequest{}
	mi := &file_follow_v1_follow_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelFollowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelFollowRequest) ProtoMessage() {}

func (x *CancelFollowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelFollowRequest.ProtoReflect.Descriptor instead.
func (*CancelFollowRequest) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{12}
}

func (x *CancelFollowRequest) GetFollowee() int64 {
	if x != nil {
		return x.Followee
	}
	return 0
}

func (x *CancelFollowRequest) GetFollower() int64 {
	if x != nil {
		return x.Follower
	}
	return 0
}

type CancelFollowResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelFollowResponse) Reset() {
	*x = CancelFollowResponse{}
	mi := &file_follow_v1_follow_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelFollowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelFollowResponse) ProtoMessage() {}

func (x *CancelFollowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_follow_v1_follow_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelFollowResponse.ProtoReflect.Descriptor instead.
func (*CancelFollowResponse) Descriptor() ([]byte, []int) {
	return file_follow_v1_follow_proto_rawDescGZIP(), []int{13}
}

var File_follow_v1_follow_proto protoreflect.FileDescriptor

const file_follow_v1_follow_proto_rawDesc = "" +
	"\n" +
	"\x16follow/v1/follow.proto\x12\tfollow.v1\"0\n" +
	"\x12GetFollowerRequest\x12\x1a\n" +
	"\bfollowee\x18\x01 \x01(\x03R\bfollowee\"[\n" +
	"\x13GetFollowerResponse\x12D\n" +
	"\x10follow_relations\x18\x01 \x03(\v2\x19.follow.v1.FollowRelationR\x0ffollowRelations\"J\n" +
	"\fFollowStatic\x12\x1c\n" +
	"\tfollowers\x18\x01 \x01(\x03R\tfollowers\x12\x1c\n" +
	"\tfollowees\x18\x02 \x01(\x03R\tfollowees\"4\n" +
	"\x16GetFollowStaticRequest\x12\x1a\n" +
	"\bfollowee\x18\x01 \x01(\x03R\bfollowee\"V\n" +
	"\x17GetFollowStaticResponse\x12;\n" +
	"\ffollowStatic\x18\x01 \x01(\v2\x17.follow.v1.FollowStaticR\ffollowStatic\"X\n" +
	"\x0eFollowRelation\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x1a\n" +
	"\bfollowee\x18\x02 \x01(\x03R\bfollowee\x12\x1a\n" +
	"\bfollower\x18\x03 \x01(\x03R\bfollower\"^\n" +
	"\x12GetFolloweeRequest\x12\x1a\n" +
	"\bfollower\x18\x01 \x01(\x03R\bfollower\x12\x16\n" +
	"\x06offset\x18\x02 \x01(\x03R\x06offset\x12\x14\n" +
	"\x05limit\x18\x03 \x01(\x03R\x05limit\"[\n" +
	"\x13GetFolloweeResponse\x12D\n" +
	"\x10follow_relations\x18\x01 \x03(\v2\x19.follow.v1.FollowRelationR\x0ffollowRelations\"K\n" +
	"\x11FollowInfoRequest\x12\x1a\n" +
	"\bfollower\x18\x01 \x01(\x03R\bfollower\x12\x1a\n" +
	"\bfollowee\x18\x02 \x01(\x03R\bfollowee\"X\n" +
	"\x12FollowInfoResponse\x12B\n" +
	"\x0ffollow_relation\x18\x01 \x01(\v2\x19.follow.v1.FollowRelationR\x0efollowRelation\"G\n" +
	"\rFollowRequest\x12\x1a\n" +
	"\bfollowee\x18\x01 \x01(\x03R\bfollowee\x12\x1a\n" +
	"\bfollower\x18\x02 \x01(\x03R\bfollower\"\x10\n" +
	"\x0eFollowResponse\"M\n" +
	"\x13CancelFollowRequest\x12\x1a\n" +
	"\bfollowee\x18\x01 \x01(\x03R\bfollowee\x12\x1a\n" +
	"\bfollower\x18\x02 \x01(\x03R\bfollower\"\x16\n" +
	"\x14CancelFollowResponse2\xe0\x03\n" +
	"\rFollowService\x12L\n" +
	"\vGetFollowee\x12\x1d.follow.v1.GetFolloweeRequest\x1a\x1e.follow.v1.GetFolloweeResponse\x12I\n" +
	"\n" +
	"FollowInfo\x12\x1c.follow.v1.FollowInfoRequest\x1a\x1d.follow.v1.FollowInfoResponse\x12=\n" +
	"\x06Follow\x12\x18.follow.v1.FollowRequest\x1a\x19.follow.v1.FollowResponse\x12O\n" +
	"\fCancelFollow\x12\x1e.follow.v1.CancelFollowRequest\x1a\x1f.follow.v1.CancelFollowResponse\x12X\n" +
	"\x0fGetFollowStatic\x12!.follow.v1.GetFollowStaticRequest\x1a\".follow.v1.GetFollowStaticResponse\x12L\n" +
	"\vGetFollower\x12\x1d.follow.v1.GetFollowerRequest\x1a\x1e.follow.v1.GetFollowerResponseB\x8a\x01\n" +
	"\rcom.follow.v1B\vFollowProtoP\x01Z'webook/api/proto/gen/follow/v1;followv1\xa2\x02\x03FXX\xaa\x02\tFollow.V1\xca\x02\tFollow\\V1\xe2\x02\x15Follow\\V1\\GPBMetadata\xea\x02\n" +
	"Follow::V1b\x06proto3"

var (
	file_follow_v1_follow_proto_rawDescOnce sync.Once
	file_follow_v1_follow_proto_rawDescData []byte
)

func file_follow_v1_follow_proto_rawDescGZIP() []byte {
	file_follow_v1_follow_proto_rawDescOnce.Do(func() {
		file_follow_v1_follow_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_follow_v1_follow_proto_rawDesc), len(file_follow_v1_follow_proto_rawDesc)))
	})
	return file_follow_v1_follow_proto_rawDescData
}

var file_follow_v1_follow_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_follow_v1_follow_proto_goTypes = []any{
	(*GetFollowerRequest)(nil),      // 0: follow.v1.GetFollowerRequest
	(*GetFollowerResponse)(nil),     // 1: follow.v1.GetFollowerResponse
	(*FollowStatic)(nil),            // 2: follow.v1.FollowStatic
	(*GetFollowStaticRequest)(nil),  // 3: follow.v1.GetFollowStaticRequest
	(*GetFollowStaticResponse)(nil), // 4: follow.v1.GetFollowStaticResponse
	(*FollowRelation)(nil),          // 5: follow.v1.FollowRelation
	(*GetFolloweeRequest)(nil),      // 6: follow.v1.GetFolloweeRequest
	(*GetFolloweeResponse)(nil),     // 7: follow.v1.GetFolloweeResponse
	(*FollowInfoRequest)(nil),       // 8: follow.v1.FollowInfoRequest
	(*FollowInfoResponse)(nil),      // 9: follow.v1.FollowInfoResponse
	(*FollowRequest)(nil),           // 10: follow.v1.FollowRequest
	(*FollowResponse)(nil),          // 11: follow.v1.FollowResponse
	(*CancelFollowRequest)(nil),     // 12: follow.v1.CancelFollowRequest
	(*CancelFollowResponse)(nil),    // 13: follow.v1.CancelFollowResponse
}
var file_follow_v1_follow_proto_depIdxs = []int32{
	5,  // 0: follow.v1.GetFollowerResponse.follow_relations:type_name -> follow.v1.FollowRelation
	2,  // 1: follow.v1.GetFollowStaticResponse.followStatic:type_name -> follow.v1.FollowStatic
	5,  // 2: follow.v1.GetFolloweeResponse.follow_relations:type_name -> follow.v1.FollowRelation
	5,  // 3: follow.v1.FollowInfoResponse.follow_relation:type_name -> follow.v1.FollowRelation
	6,  // 4: follow.v1.FollowService.GetFollowee:input_type -> follow.v1.GetFolloweeRequest
	8,  // 5: follow.v1.FollowService.FollowInfo:input_type -> follow.v1.FollowInfoRequest
	10, // 6: follow.v1.FollowService.Follow:input_type -> follow.v1.FollowRequest
	12, // 7: follow.v1.FollowService.CancelFollow:input_type -> follow.v1.CancelFollowRequest
	3,  // 8: follow.v1.FollowService.GetFollowStatic:input_type -> follow.v1.GetFollowStaticRequest
	0,  // 9: follow.v1.FollowService.GetFollower:input_type -> follow.v1.GetFollowerRequest
	7,  // 10: follow.v1.FollowService.GetFollowee:output_type -> follow.v1.GetFolloweeResponse
	9,  // 11: follow.v1.FollowService.FollowInfo:output_type -> follow.v1.FollowInfoResponse
	11, // 12: follow.v1.FollowService.Follow:output_type -> follow.v1.FollowResponse
	13, // 13: follow.v1.FollowService.CancelFollow:output_type -> follow.v1.CancelFollowResponse
	4,  // 14: follow.v1.FollowService.GetFollowStatic:output_type -> follow.v1.GetFollowStaticResponse
	1,  // 15: follow.v1.FollowService.GetFollower:output_type -> follow.v1.GetFollowerResponse
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_follow_v1_follow_proto_init() }
func file_follow_v1_follow_proto_init() {
	if File_follow_v1_follow_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_follow_v1_follow_proto_rawDesc), len(file_follow_v1_follow_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_follow_v1_follow_proto_goTypes,
		DependencyIndexes: file_follow_v1_follow_proto_depIdxs,
		MessageInfos:      file_follow_v1_follow_proto_msgTypes,
	}.Build()
	File_follow_v1_follow_proto = out.File
	file_follow_v1_follow_proto_goTypes = nil
	file_follow_v1_follow_proto_depIdxs = nil
}
