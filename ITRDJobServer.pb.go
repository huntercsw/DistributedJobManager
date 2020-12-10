// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ITRDJobServer.proto

package main

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type JobInfo struct {
	AccountId            string   `protobuf:"bytes,1,opt,name=accountId,proto3" json:"accountId,omitempty"`
	AccountName          string   `protobuf:"bytes,2,opt,name=accountName,proto3" json:"accountName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *JobInfo) Reset()         { *m = JobInfo{} }
func (m *JobInfo) String() string { return proto.CompactTextString(m) }
func (*JobInfo) ProtoMessage()    {}
func (*JobInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{0}
}

func (m *JobInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_JobInfo.Unmarshal(m, b)
}
func (m *JobInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_JobInfo.Marshal(b, m, deterministic)
}
func (m *JobInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JobInfo.Merge(m, src)
}
func (m *JobInfo) XXX_Size() int {
	return xxx_messageInfo_JobInfo.Size(m)
}
func (m *JobInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_JobInfo.DiscardUnknown(m)
}

var xxx_messageInfo_JobInfo proto.InternalMessageInfo

func (m *JobInfo) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *JobInfo) GetAccountName() string {
	if m != nil {
		return m.AccountName
	}
	return ""
}

type JobStatus struct {
	AccountId            string   `protobuf:"bytes,1,opt,name=accountId,proto3" json:"accountId,omitempty"`
	AccountName          string   `protobuf:"bytes,2,opt,name=accountName,proto3" json:"accountName,omitempty"`
	Position             string   `protobuf:"bytes,3,opt,name=position,proto3" json:"position,omitempty"`
	Order                string   `protobuf:"bytes,4,opt,name=order,proto3" json:"order,omitempty"`
	Trade                string   `protobuf:"bytes,5,opt,name=trade,proto3" json:"trade,omitempty"`
	ErrCode              int32    `protobuf:"varint,6,opt,name=errCode,proto3" json:"errCode,omitempty"`
	ErrMsg               string   `protobuf:"bytes,7,opt,name=errMsg,proto3" json:"errMsg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *JobStatus) Reset()         { *m = JobStatus{} }
func (m *JobStatus) String() string { return proto.CompactTextString(m) }
func (*JobStatus) ProtoMessage()    {}
func (*JobStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{1}
}

func (m *JobStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_JobStatus.Unmarshal(m, b)
}
func (m *JobStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_JobStatus.Marshal(b, m, deterministic)
}
func (m *JobStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JobStatus.Merge(m, src)
}
func (m *JobStatus) XXX_Size() int {
	return xxx_messageInfo_JobStatus.Size(m)
}
func (m *JobStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_JobStatus.DiscardUnknown(m)
}

var xxx_messageInfo_JobStatus proto.InternalMessageInfo

func (m *JobStatus) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *JobStatus) GetAccountName() string {
	if m != nil {
		return m.AccountName
	}
	return ""
}

func (m *JobStatus) GetPosition() string {
	if m != nil {
		return m.Position
	}
	return ""
}

func (m *JobStatus) GetOrder() string {
	if m != nil {
		return m.Order
	}
	return ""
}

func (m *JobStatus) GetTrade() string {
	if m != nil {
		return m.Trade
	}
	return ""
}

func (m *JobStatus) GetErrCode() int32 {
	if m != nil {
		return m.ErrCode
	}
	return 0
}

func (m *JobStatus) GetErrMsg() string {
	if m != nil {
		return m.ErrMsg
	}
	return ""
}

type PullConfRequest struct {
	RemoteIp             string   `protobuf:"bytes,1,opt,name=remoteIp,proto3" json:"remoteIp,omitempty"`
	UserName             string   `protobuf:"bytes,2,opt,name=userName,proto3" json:"userName,omitempty"`
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	SourceConfPath       string   `protobuf:"bytes,4,opt,name=sourceConfPath,proto3" json:"sourceConfPath,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullConfRequest) Reset()         { *m = PullConfRequest{} }
func (m *PullConfRequest) String() string { return proto.CompactTextString(m) }
func (*PullConfRequest) ProtoMessage()    {}
func (*PullConfRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{2}
}

func (m *PullConfRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullConfRequest.Unmarshal(m, b)
}
func (m *PullConfRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullConfRequest.Marshal(b, m, deterministic)
}
func (m *PullConfRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullConfRequest.Merge(m, src)
}
func (m *PullConfRequest) XXX_Size() int {
	return xxx_messageInfo_PullConfRequest.Size(m)
}
func (m *PullConfRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PullConfRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PullConfRequest proto.InternalMessageInfo

func (m *PullConfRequest) GetRemoteIp() string {
	if m != nil {
		return m.RemoteIp
	}
	return ""
}

func (m *PullConfRequest) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *PullConfRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *PullConfRequest) GetSourceConfPath() string {
	if m != nil {
		return m.SourceConfPath
	}
	return ""
}

type PullConfResponse struct {
	Status               int32    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullConfResponse) Reset()         { *m = PullConfResponse{} }
func (m *PullConfResponse) String() string { return proto.CompactTextString(m) }
func (*PullConfResponse) ProtoMessage()    {}
func (*PullConfResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{3}
}

func (m *PullConfResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullConfResponse.Unmarshal(m, b)
}
func (m *PullConfResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullConfResponse.Marshal(b, m, deterministic)
}
func (m *PullConfResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullConfResponse.Merge(m, src)
}
func (m *PullConfResponse) XXX_Size() int {
	return xxx_messageInfo_PullConfResponse.Size(m)
}
func (m *PullConfResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PullConfResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PullConfResponse proto.InternalMessageInfo

func (m *PullConfResponse) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

type PushFileRequest struct {
	RemoteIp             string   `protobuf:"bytes,1,opt,name=remoteIp,proto3" json:"remoteIp,omitempty"`
	UserName             string   `protobuf:"bytes,2,opt,name=userName,proto3" json:"userName,omitempty"`
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	Dir                  string   `protobuf:"bytes,4,opt,name=dir,proto3" json:"dir,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushFileRequest) Reset()         { *m = PushFileRequest{} }
func (m *PushFileRequest) String() string { return proto.CompactTextString(m) }
func (*PushFileRequest) ProtoMessage()    {}
func (*PushFileRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{4}
}

func (m *PushFileRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushFileRequest.Unmarshal(m, b)
}
func (m *PushFileRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushFileRequest.Marshal(b, m, deterministic)
}
func (m *PushFileRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushFileRequest.Merge(m, src)
}
func (m *PushFileRequest) XXX_Size() int {
	return xxx_messageInfo_PushFileRequest.Size(m)
}
func (m *PushFileRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PushFileRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PushFileRequest proto.InternalMessageInfo

func (m *PushFileRequest) GetRemoteIp() string {
	if m != nil {
		return m.RemoteIp
	}
	return ""
}

func (m *PushFileRequest) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *PushFileRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *PushFileRequest) GetDir() string {
	if m != nil {
		return m.Dir
	}
	return ""
}

type PushFileResponse struct {
	Status               int32    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushFileResponse) Reset()         { *m = PushFileResponse{} }
func (m *PushFileResponse) String() string { return proto.CompactTextString(m) }
func (*PushFileResponse) ProtoMessage()    {}
func (*PushFileResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{5}
}

func (m *PushFileResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushFileResponse.Unmarshal(m, b)
}
func (m *PushFileResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushFileResponse.Marshal(b, m, deterministic)
}
func (m *PushFileResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushFileResponse.Merge(m, src)
}
func (m *PushFileResponse) XXX_Size() int {
	return xxx_messageInfo_PushFileResponse.Size(m)
}
func (m *PushFileResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PushFileResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PushFileResponse proto.InternalMessageInfo

func (m *PushFileResponse) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

type AliveCheckRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AliveCheckRequest) Reset()         { *m = AliveCheckRequest{} }
func (m *AliveCheckRequest) String() string { return proto.CompactTextString(m) }
func (*AliveCheckRequest) ProtoMessage()    {}
func (*AliveCheckRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{6}
}

func (m *AliveCheckRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AliveCheckRequest.Unmarshal(m, b)
}
func (m *AliveCheckRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AliveCheckRequest.Marshal(b, m, deterministic)
}
func (m *AliveCheckRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AliveCheckRequest.Merge(m, src)
}
func (m *AliveCheckRequest) XXX_Size() int {
	return xxx_messageInfo_AliveCheckRequest.Size(m)
}
func (m *AliveCheckRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AliveCheckRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AliveCheckRequest proto.InternalMessageInfo

type AliveCheckResponse struct {
	Status               int32    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AliveCheckResponse) Reset()         { *m = AliveCheckResponse{} }
func (m *AliveCheckResponse) String() string { return proto.CompactTextString(m) }
func (*AliveCheckResponse) ProtoMessage()    {}
func (*AliveCheckResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b8cde854a210f0e, []int{7}
}

func (m *AliveCheckResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AliveCheckResponse.Unmarshal(m, b)
}
func (m *AliveCheckResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AliveCheckResponse.Marshal(b, m, deterministic)
}
func (m *AliveCheckResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AliveCheckResponse.Merge(m, src)
}
func (m *AliveCheckResponse) XXX_Size() int {
	return xxx_messageInfo_AliveCheckResponse.Size(m)
}
func (m *AliveCheckResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AliveCheckResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AliveCheckResponse proto.InternalMessageInfo

func (m *AliveCheckResponse) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func init() {
	proto.RegisterType((*JobInfo)(nil), "main.JobInfo")
	proto.RegisterType((*JobStatus)(nil), "main.JobStatus")
	proto.RegisterType((*PullConfRequest)(nil), "main.PullConfRequest")
	proto.RegisterType((*PullConfResponse)(nil), "main.PullConfResponse")
	proto.RegisterType((*PushFileRequest)(nil), "main.PushFileRequest")
	proto.RegisterType((*PushFileResponse)(nil), "main.PushFileResponse")
	proto.RegisterType((*AliveCheckRequest)(nil), "main.AliveCheckRequest")
	proto.RegisterType((*AliveCheckResponse)(nil), "main.AliveCheckResponse")
}

func init() { proto.RegisterFile("ITRDJobServer.proto", fileDescriptor_1b8cde854a210f0e) }

var fileDescriptor_1b8cde854a210f0e = []byte{
	// 434 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0xdb, 0x26, 0x69, 0x06, 0x41, 0xcb, 0x14, 0xca, 0x2a, 0xe2, 0x10, 0xed, 0x01, 0x55,
	0x05, 0xe5, 0x00, 0x47, 0xc4, 0xa1, 0x0a, 0x42, 0x72, 0xa4, 0xa2, 0xca, 0xe5, 0x05, 0xd6, 0xf1,
	0x34, 0xb1, 0x48, 0x3c, 0x66, 0x77, 0xdd, 0x3e, 0x05, 0x6f, 0xc1, 0xab, 0xf0, 0x5e, 0x68, 0xd7,
	0x6b, 0x3b, 0x4d, 0x25, 0x40, 0x42, 0xdc, 0xfc, 0x7d, 0xf3, 0xf7, 0xcd, 0xe7, 0x59, 0x38, 0x89,
	0xbf, 0x24, 0x1f, 0xe7, 0x9c, 0x5e, 0x93, 0xbe, 0x25, 0x3d, 0x2d, 0x35, 0x5b, 0xc6, 0x83, 0x8d,
	0xca, 0x0b, 0x19, 0xc3, 0x70, 0xce, 0x69, 0x5c, 0xdc, 0x30, 0xbe, 0x84, 0x91, 0x5a, 0x2c, 0xb8,
	0x2a, 0x6c, 0x9c, 0x89, 0x68, 0x12, 0x9d, 0x8d, 0x92, 0x8e, 0xc0, 0x09, 0x3c, 0x0a, 0xe0, 0xb3,
	0xda, 0x90, 0xd8, 0xf3, 0xf1, 0x6d, 0x4a, 0xfe, 0x8c, 0x60, 0xe4, 0x86, 0x58, 0x65, 0x2b, 0xf3,
	0xaf, 0xdd, 0x70, 0x0c, 0x87, 0x25, 0x9b, 0xdc, 0xe6, 0x5c, 0x88, 0x7d, 0x1f, 0x6e, 0x31, 0x3e,
	0x83, 0x3e, 0xeb, 0x8c, 0xb4, 0x38, 0xf0, 0x81, 0x1a, 0x38, 0xd6, 0x6a, 0x95, 0x91, 0xe8, 0xd7,
	0xac, 0x07, 0x28, 0x60, 0x48, 0x5a, 0xcf, 0x38, 0x23, 0x31, 0x98, 0x44, 0x67, 0xfd, 0xa4, 0x81,
	0x78, 0x0a, 0x03, 0xd2, 0xfa, 0xd2, 0x2c, 0xc5, 0xd0, 0x17, 0x04, 0x24, 0xbf, 0x47, 0x70, 0x74,
	0x55, 0xad, 0xd7, 0x33, 0x2e, 0x6e, 0x12, 0xfa, 0x56, 0x91, 0xb1, 0x4e, 0x8d, 0xa6, 0x0d, 0x5b,
	0x8a, 0xcb, 0xb0, 0x4c, 0x8b, 0x5d, 0xac, 0x32, 0xa4, 0xb7, 0x16, 0x69, 0xb1, 0xdf, 0x42, 0x19,
	0x73, 0xc7, 0x3a, 0x6b, 0xb7, 0x08, 0x18, 0x5f, 0xc1, 0x13, 0xc3, 0x95, 0x5e, 0x90, 0x1b, 0x74,
	0xa5, 0xec, 0x2a, 0xac, 0xb3, 0xc3, 0xca, 0x73, 0x38, 0xee, 0xe4, 0x98, 0x92, 0x0b, 0xe3, 0xb5,
	0x1b, 0xef, 0xb3, 0x57, 0xd3, 0x4f, 0x02, 0x92, 0x77, 0x4e, 0xba, 0x59, 0x7d, 0xca, 0xd7, 0xf4,
	0x3f, 0xa5, 0x1f, 0xc3, 0x7e, 0x96, 0x37, 0xf6, 0xbb, 0xcf, 0x5a, 0x64, 0x33, 0xf8, 0x0f, 0x22,
	0x4f, 0xe0, 0xe9, 0xc5, 0x3a, 0xbf, 0xa5, 0xd9, 0x8a, 0x16, 0x5f, 0x83, 0x4c, 0xf9, 0x06, 0x70,
	0x9b, 0xfc, 0x7d, 0x8b, 0xb7, 0x3f, 0xf6, 0x00, 0xe6, 0x9c, 0x5e, 0xaa, 0x42, 0x2d, 0x49, 0xe3,
	0x39, 0x0c, 0x92, 0xaa, 0x98, 0x73, 0x8a, 0x8f, 0xa7, 0xee, 0xac, 0xa7, 0xe1, 0xa6, 0xc7, 0x47,
	0x2d, 0xac, 0xcf, 0x52, 0xf6, 0xf0, 0x35, 0x0c, 0xaf, 0x2d, 0x97, 0x7f, 0x97, 0xfc, 0x01, 0xa0,
	0xf1, 0x3e, 0x5f, 0xe2, 0xf3, 0x3a, 0x61, 0xe7, 0x38, 0xc6, 0xa7, 0xbb, 0x74, 0x2d, 0x5e, 0xf6,
	0xf0, 0x3d, 0x1c, 0x36, 0xae, 0x74, 0xc5, 0xf7, 0x7e, 0x4f, 0x57, 0x7c, 0xdf, 0x3c, 0xd9, 0xc3,
	0x0b, 0x80, 0xce, 0x11, 0x7c, 0x51, 0xe7, 0x3d, 0x30, 0x6e, 0x2c, 0x1e, 0x06, 0x9a, 0x16, 0xe9,
	0xc0, 0x3f, 0xf5, 0x77, 0xbf, 0x02, 0x00, 0x00, 0xff, 0xff, 0xa1, 0x46, 0xb9, 0x17, 0x01, 0x04,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// JobManagerClient is the client API for JobManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type JobManagerClient interface {
	RunJob(ctx context.Context, in *JobInfo, opts ...grpc.CallOption) (*JobStatus, error)
	StopJob(ctx context.Context, in *JobInfo, opts ...grpc.CallOption) (*JobStatus, error)
	PullConfig(ctx context.Context, in *PullConfRequest, opts ...grpc.CallOption) (*PullConfResponse, error)
	PushFile(ctx context.Context, in *PushFileRequest, opts ...grpc.CallOption) (*PushFileResponse, error)
	AliveCheck(ctx context.Context, in *AliveCheckRequest, opts ...grpc.CallOption) (*AliveCheckResponse, error)
}

type jobManagerClient struct {
	cc *grpc.ClientConn
}

func NewJobManagerClient(cc *grpc.ClientConn) JobManagerClient {
	return &jobManagerClient{cc}
}

func (c *jobManagerClient) RunJob(ctx context.Context, in *JobInfo, opts ...grpc.CallOption) (*JobStatus, error) {
	out := new(JobStatus)
	err := c.cc.Invoke(ctx, "/main.JobManager/RunJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobManagerClient) StopJob(ctx context.Context, in *JobInfo, opts ...grpc.CallOption) (*JobStatus, error) {
	out := new(JobStatus)
	err := c.cc.Invoke(ctx, "/main.JobManager/StopJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobManagerClient) PullConfig(ctx context.Context, in *PullConfRequest, opts ...grpc.CallOption) (*PullConfResponse, error) {
	out := new(PullConfResponse)
	err := c.cc.Invoke(ctx, "/main.JobManager/PullConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobManagerClient) PushFile(ctx context.Context, in *PushFileRequest, opts ...grpc.CallOption) (*PushFileResponse, error) {
	out := new(PushFileResponse)
	err := c.cc.Invoke(ctx, "/main.JobManager/PushFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobManagerClient) AliveCheck(ctx context.Context, in *AliveCheckRequest, opts ...grpc.CallOption) (*AliveCheckResponse, error) {
	out := new(AliveCheckResponse)
	err := c.cc.Invoke(ctx, "/main.JobManager/AliveCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JobManagerServer is the server API for JobManager service.
type JobManagerServer interface {
	RunJob(context.Context, *JobInfo) (*JobStatus, error)
	StopJob(context.Context, *JobInfo) (*JobStatus, error)
	PullConfig(context.Context, *PullConfRequest) (*PullConfResponse, error)
	PushFile(context.Context, *PushFileRequest) (*PushFileResponse, error)
	AliveCheck(context.Context, *AliveCheckRequest) (*AliveCheckResponse, error)
}

// UnimplementedJobManagerServer can be embedded to have forward compatible implementations.
type UnimplementedJobManagerServer struct {
}

func (*UnimplementedJobManagerServer) RunJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunJob not implemented")
}
func (*UnimplementedJobManagerServer) StopJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopJob not implemented")
}
func (*UnimplementedJobManagerServer) PullConfig(ctx context.Context, req *PullConfRequest) (*PullConfResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PullConfig not implemented")
}
func (*UnimplementedJobManagerServer) PushFile(ctx context.Context, req *PushFileRequest) (*PushFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushFile not implemented")
}
func (*UnimplementedJobManagerServer) AliveCheck(ctx context.Context, req *AliveCheckRequest) (*AliveCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AliveCheck not implemented")
}

func RegisterJobManagerServer(s *grpc.Server, srv JobManagerServer) {
	s.RegisterService(&_JobManager_serviceDesc, srv)
}

func _JobManager_RunJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobManagerServer).RunJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.JobManager/RunJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobManagerServer).RunJob(ctx, req.(*JobInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobManager_StopJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobManagerServer).StopJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.JobManager/StopJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobManagerServer).StopJob(ctx, req.(*JobInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobManager_PullConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PullConfRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobManagerServer).PullConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.JobManager/PullConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobManagerServer).PullConfig(ctx, req.(*PullConfRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobManager_PushFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobManagerServer).PushFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.JobManager/PushFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobManagerServer).PushFile(ctx, req.(*PushFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobManager_AliveCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AliveCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobManagerServer).AliveCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.JobManager/AliveCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobManagerServer).AliveCheck(ctx, req.(*AliveCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _JobManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.JobManager",
	HandlerType: (*JobManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RunJob",
			Handler:    _JobManager_RunJob_Handler,
		},
		{
			MethodName: "StopJob",
			Handler:    _JobManager_StopJob_Handler,
		},
		{
			MethodName: "PullConfig",
			Handler:    _JobManager_PullConfig_Handler,
		},
		{
			MethodName: "PushFile",
			Handler:    _JobManager_PushFile_Handler,
		},
		{
			MethodName: "AliveCheck",
			Handler:    _JobManager_AliveCheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ITRDJobServer.proto",
}
