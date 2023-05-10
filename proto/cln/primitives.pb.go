// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v4.22.3
// source: proto/cln/primitives.proto

package cln

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ChannelSide int32

const (
	ChannelSide_LOCAL  ChannelSide = 0
	ChannelSide_REMOTE ChannelSide = 1
)

// Enum value maps for ChannelSide.
var (
	ChannelSide_name = map[int32]string{
		0: "LOCAL",
		1: "REMOTE",
	}
	ChannelSide_value = map[string]int32{
		"LOCAL":  0,
		"REMOTE": 1,
	}
)

func (x ChannelSide) Enum() *ChannelSide {
	p := new(ChannelSide)
	*p = x
	return p
}

func (x ChannelSide) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ChannelSide) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_cln_primitives_proto_enumTypes[0].Descriptor()
}

func (ChannelSide) Type() protoreflect.EnumType {
	return &file_proto_cln_primitives_proto_enumTypes[0]
}

func (x ChannelSide) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ChannelSide.Descriptor instead.
func (ChannelSide) EnumDescriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{0}
}

type ChannelState int32

const (
	ChannelState_Openingd                ChannelState = 0
	ChannelState_ChanneldAwaitingLockin  ChannelState = 1
	ChannelState_ChanneldNormal          ChannelState = 2
	ChannelState_ChanneldShuttingDown    ChannelState = 3
	ChannelState_ClosingdSigexchange     ChannelState = 4
	ChannelState_ClosingdComplete        ChannelState = 5
	ChannelState_AwaitingUnilateral      ChannelState = 6
	ChannelState_FundingSpendSeen        ChannelState = 7
	ChannelState_Onchain                 ChannelState = 8
	ChannelState_DualopendOpenInit       ChannelState = 9
	ChannelState_DualopendAwaitingLockin ChannelState = 10
)

// Enum value maps for ChannelState.
var (
	ChannelState_name = map[int32]string{
		0:  "Openingd",
		1:  "ChanneldAwaitingLockin",
		2:  "ChanneldNormal",
		3:  "ChanneldShuttingDown",
		4:  "ClosingdSigexchange",
		5:  "ClosingdComplete",
		6:  "AwaitingUnilateral",
		7:  "FundingSpendSeen",
		8:  "Onchain",
		9:  "DualopendOpenInit",
		10: "DualopendAwaitingLockin",
	}
	ChannelState_value = map[string]int32{
		"Openingd":                0,
		"ChanneldAwaitingLockin":  1,
		"ChanneldNormal":          2,
		"ChanneldShuttingDown":    3,
		"ClosingdSigexchange":     4,
		"ClosingdComplete":        5,
		"AwaitingUnilateral":      6,
		"FundingSpendSeen":        7,
		"Onchain":                 8,
		"DualopendOpenInit":       9,
		"DualopendAwaitingLockin": 10,
	}
)

func (x ChannelState) Enum() *ChannelState {
	p := new(ChannelState)
	*p = x
	return p
}

func (x ChannelState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ChannelState) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_cln_primitives_proto_enumTypes[1].Descriptor()
}

func (ChannelState) Type() protoreflect.EnumType {
	return &file_proto_cln_primitives_proto_enumTypes[1]
}

func (x ChannelState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ChannelState.Descriptor instead.
func (ChannelState) EnumDescriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{1}
}

type HtlcState int32

const (
	HtlcState_SentAddHtlc             HtlcState = 0
	HtlcState_SentAddCommit           HtlcState = 1
	HtlcState_RcvdAddRevocation       HtlcState = 2
	HtlcState_RcvdAddAckCommit        HtlcState = 3
	HtlcState_SentAddAckRevocation    HtlcState = 4
	HtlcState_RcvdAddAckRevocation    HtlcState = 5
	HtlcState_RcvdRemoveHtlc          HtlcState = 6
	HtlcState_RcvdRemoveCommit        HtlcState = 7
	HtlcState_SentRemoveRevocation    HtlcState = 8
	HtlcState_SentRemoveAckCommit     HtlcState = 9
	HtlcState_RcvdRemoveAckRevocation HtlcState = 10
)

// Enum value maps for HtlcState.
var (
	HtlcState_name = map[int32]string{
		0:  "SentAddHtlc",
		1:  "SentAddCommit",
		2:  "RcvdAddRevocation",
		3:  "RcvdAddAckCommit",
		4:  "SentAddAckRevocation",
		5:  "RcvdAddAckRevocation",
		6:  "RcvdRemoveHtlc",
		7:  "RcvdRemoveCommit",
		8:  "SentRemoveRevocation",
		9:  "SentRemoveAckCommit",
		10: "RcvdRemoveAckRevocation",
	}
	HtlcState_value = map[string]int32{
		"SentAddHtlc":             0,
		"SentAddCommit":           1,
		"RcvdAddRevocation":       2,
		"RcvdAddAckCommit":        3,
		"SentAddAckRevocation":    4,
		"RcvdAddAckRevocation":    5,
		"RcvdRemoveHtlc":          6,
		"RcvdRemoveCommit":        7,
		"SentRemoveRevocation":    8,
		"SentRemoveAckCommit":     9,
		"RcvdRemoveAckRevocation": 10,
	}
)

func (x HtlcState) Enum() *HtlcState {
	p := new(HtlcState)
	*p = x
	return p
}

func (x HtlcState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HtlcState) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_cln_primitives_proto_enumTypes[2].Descriptor()
}

func (HtlcState) Type() protoreflect.EnumType {
	return &file_proto_cln_primitives_proto_enumTypes[2]
}

func (x HtlcState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HtlcState.Descriptor instead.
func (HtlcState) EnumDescriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{2}
}

type Amount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msat uint64 `protobuf:"varint,1,opt,name=msat,proto3" json:"msat,omitempty"`
}

func (x *Amount) Reset() {
	*x = Amount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Amount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Amount) ProtoMessage() {}

func (x *Amount) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Amount.ProtoReflect.Descriptor instead.
func (*Amount) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{0}
}

func (x *Amount) GetMsat() uint64 {
	if x != nil {
		return x.Msat
	}
	return 0
}

type AmountOrAll struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//
	//	*AmountOrAll_Amount
	//	*AmountOrAll_All
	Value isAmountOrAll_Value `protobuf_oneof:"value"`
}

func (x *AmountOrAll) Reset() {
	*x = AmountOrAll{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AmountOrAll) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AmountOrAll) ProtoMessage() {}

func (x *AmountOrAll) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AmountOrAll.ProtoReflect.Descriptor instead.
func (*AmountOrAll) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{1}
}

func (m *AmountOrAll) GetValue() isAmountOrAll_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *AmountOrAll) GetAmount() *Amount {
	if x, ok := x.GetValue().(*AmountOrAll_Amount); ok {
		return x.Amount
	}
	return nil
}

func (x *AmountOrAll) GetAll() bool {
	if x, ok := x.GetValue().(*AmountOrAll_All); ok {
		return x.All
	}
	return false
}

type isAmountOrAll_Value interface {
	isAmountOrAll_Value()
}

type AmountOrAll_Amount struct {
	Amount *Amount `protobuf:"bytes,1,opt,name=amount,proto3,oneof"`
}

type AmountOrAll_All struct {
	All bool `protobuf:"varint,2,opt,name=all,proto3,oneof"`
}

func (*AmountOrAll_Amount) isAmountOrAll_Value() {}

func (*AmountOrAll_All) isAmountOrAll_Value() {}

type AmountOrAny struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//
	//	*AmountOrAny_Amount
	//	*AmountOrAny_Any
	Value isAmountOrAny_Value `protobuf_oneof:"value"`
}

func (x *AmountOrAny) Reset() {
	*x = AmountOrAny{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AmountOrAny) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AmountOrAny) ProtoMessage() {}

func (x *AmountOrAny) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AmountOrAny.ProtoReflect.Descriptor instead.
func (*AmountOrAny) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{2}
}

func (m *AmountOrAny) GetValue() isAmountOrAny_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *AmountOrAny) GetAmount() *Amount {
	if x, ok := x.GetValue().(*AmountOrAny_Amount); ok {
		return x.Amount
	}
	return nil
}

func (x *AmountOrAny) GetAny() bool {
	if x, ok := x.GetValue().(*AmountOrAny_Any); ok {
		return x.Any
	}
	return false
}

type isAmountOrAny_Value interface {
	isAmountOrAny_Value()
}

type AmountOrAny_Amount struct {
	Amount *Amount `protobuf:"bytes,1,opt,name=amount,proto3,oneof"`
}

type AmountOrAny_Any struct {
	Any bool `protobuf:"varint,2,opt,name=any,proto3,oneof"`
}

func (*AmountOrAny_Amount) isAmountOrAny_Value() {}

func (*AmountOrAny_Any) isAmountOrAny_Value() {}

type ChannelStateChangeCause struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ChannelStateChangeCause) Reset() {
	*x = ChannelStateChangeCause{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChannelStateChangeCause) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChannelStateChangeCause) ProtoMessage() {}

func (x *ChannelStateChangeCause) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChannelStateChangeCause.ProtoReflect.Descriptor instead.
func (*ChannelStateChangeCause) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{3}
}

type Outpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Txid   []byte `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	Outnum uint32 `protobuf:"varint,2,opt,name=outnum,proto3" json:"outnum,omitempty"`
}

func (x *Outpoint) Reset() {
	*x = Outpoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Outpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Outpoint) ProtoMessage() {}

func (x *Outpoint) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Outpoint.ProtoReflect.Descriptor instead.
func (*Outpoint) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{4}
}

func (x *Outpoint) GetTxid() []byte {
	if x != nil {
		return x.Txid
	}
	return nil
}

func (x *Outpoint) GetOutnum() uint32 {
	if x != nil {
		return x.Outnum
	}
	return 0
}

type Feerate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Style:
	//
	//	*Feerate_Slow
	//	*Feerate_Normal
	//	*Feerate_Urgent
	//	*Feerate_Perkb
	//	*Feerate_Perkw
	Style isFeerate_Style `protobuf_oneof:"style"`
}

func (x *Feerate) Reset() {
	*x = Feerate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feerate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feerate) ProtoMessage() {}

func (x *Feerate) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feerate.ProtoReflect.Descriptor instead.
func (*Feerate) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{5}
}

func (m *Feerate) GetStyle() isFeerate_Style {
	if m != nil {
		return m.Style
	}
	return nil
}

func (x *Feerate) GetSlow() bool {
	if x, ok := x.GetStyle().(*Feerate_Slow); ok {
		return x.Slow
	}
	return false
}

func (x *Feerate) GetNormal() bool {
	if x, ok := x.GetStyle().(*Feerate_Normal); ok {
		return x.Normal
	}
	return false
}

func (x *Feerate) GetUrgent() bool {
	if x, ok := x.GetStyle().(*Feerate_Urgent); ok {
		return x.Urgent
	}
	return false
}

func (x *Feerate) GetPerkb() uint32 {
	if x, ok := x.GetStyle().(*Feerate_Perkb); ok {
		return x.Perkb
	}
	return 0
}

func (x *Feerate) GetPerkw() uint32 {
	if x, ok := x.GetStyle().(*Feerate_Perkw); ok {
		return x.Perkw
	}
	return 0
}

type isFeerate_Style interface {
	isFeerate_Style()
}

type Feerate_Slow struct {
	Slow bool `protobuf:"varint,1,opt,name=slow,proto3,oneof"`
}

type Feerate_Normal struct {
	Normal bool `protobuf:"varint,2,opt,name=normal,proto3,oneof"`
}

type Feerate_Urgent struct {
	Urgent bool `protobuf:"varint,3,opt,name=urgent,proto3,oneof"`
}

type Feerate_Perkb struct {
	Perkb uint32 `protobuf:"varint,4,opt,name=perkb,proto3,oneof"`
}

type Feerate_Perkw struct {
	Perkw uint32 `protobuf:"varint,5,opt,name=perkw,proto3,oneof"`
}

func (*Feerate_Slow) isFeerate_Style() {}

func (*Feerate_Normal) isFeerate_Style() {}

func (*Feerate_Urgent) isFeerate_Style() {}

func (*Feerate_Perkb) isFeerate_Style() {}

func (*Feerate_Perkw) isFeerate_Style() {}

type OutputDesc struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string  `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Amount  *Amount `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *OutputDesc) Reset() {
	*x = OutputDesc{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OutputDesc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutputDesc) ProtoMessage() {}

func (x *OutputDesc) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutputDesc.ProtoReflect.Descriptor instead.
func (*OutputDesc) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{6}
}

func (x *OutputDesc) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *OutputDesc) GetAmount() *Amount {
	if x != nil {
		return x.Amount
	}
	return nil
}

type RouteHop struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id             []byte  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ShortChannelId string  `protobuf:"bytes,2,opt,name=short_channel_id,json=shortChannelId,proto3" json:"short_channel_id,omitempty"`
	Feebase        *Amount `protobuf:"bytes,3,opt,name=feebase,proto3" json:"feebase,omitempty"`
	Feeprop        uint32  `protobuf:"varint,4,opt,name=feeprop,proto3" json:"feeprop,omitempty"`
	Expirydelta    uint32  `protobuf:"varint,5,opt,name=expirydelta,proto3" json:"expirydelta,omitempty"`
}

func (x *RouteHop) Reset() {
	*x = RouteHop{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RouteHop) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouteHop) ProtoMessage() {}

func (x *RouteHop) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouteHop.ProtoReflect.Descriptor instead.
func (*RouteHop) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{7}
}

func (x *RouteHop) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *RouteHop) GetShortChannelId() string {
	if x != nil {
		return x.ShortChannelId
	}
	return ""
}

func (x *RouteHop) GetFeebase() *Amount {
	if x != nil {
		return x.Feebase
	}
	return nil
}

func (x *RouteHop) GetFeeprop() uint32 {
	if x != nil {
		return x.Feeprop
	}
	return 0
}

func (x *RouteHop) GetExpirydelta() uint32 {
	if x != nil {
		return x.Expirydelta
	}
	return 0
}

type Routehint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hops []*RouteHop `protobuf:"bytes,1,rep,name=hops,proto3" json:"hops,omitempty"`
}

func (x *Routehint) Reset() {
	*x = Routehint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Routehint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Routehint) ProtoMessage() {}

func (x *Routehint) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Routehint.ProtoReflect.Descriptor instead.
func (*Routehint) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{8}
}

func (x *Routehint) GetHops() []*RouteHop {
	if x != nil {
		return x.Hops
	}
	return nil
}

type RoutehintList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hints []*Routehint `protobuf:"bytes,2,rep,name=hints,proto3" json:"hints,omitempty"`
}

func (x *RoutehintList) Reset() {
	*x = RoutehintList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RoutehintList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RoutehintList) ProtoMessage() {}

func (x *RoutehintList) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RoutehintList.ProtoReflect.Descriptor instead.
func (*RoutehintList) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{9}
}

func (x *RoutehintList) GetHints() []*Routehint {
	if x != nil {
		return x.Hints
	}
	return nil
}

type TlvEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type  uint64 `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *TlvEntry) Reset() {
	*x = TlvEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TlvEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TlvEntry) ProtoMessage() {}

func (x *TlvEntry) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TlvEntry.ProtoReflect.Descriptor instead.
func (*TlvEntry) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{10}
}

func (x *TlvEntry) GetType() uint64 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *TlvEntry) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type TlvStream struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entries []*TlvEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *TlvStream) Reset() {
	*x = TlvStream{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_cln_primitives_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TlvStream) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TlvStream) ProtoMessage() {}

func (x *TlvStream) ProtoReflect() protoreflect.Message {
	mi := &file_proto_cln_primitives_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TlvStream.ProtoReflect.Descriptor instead.
func (*TlvStream) Descriptor() ([]byte, []int) {
	return file_proto_cln_primitives_proto_rawDescGZIP(), []int{11}
}

func (x *TlvStream) GetEntries() []*TlvEntry {
	if x != nil {
		return x.Entries
	}
	return nil
}

var File_proto_cln_primitives_proto protoreflect.FileDescriptor

var file_proto_cln_primitives_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6c, 0x6e, 0x2f, 0x70, 0x72, 0x69, 0x6d,
	0x69, 0x74, 0x69, 0x76, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x63, 0x6c,
	0x6e, 0x22, 0x1c, 0x0a, 0x06, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6d,
	0x73, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x6d, 0x73, 0x61, 0x74, 0x22,
	0x51, 0x0a, 0x0b, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x4f, 0x72, 0x41, 0x6c, 0x6c, 0x12, 0x25,
	0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b,
	0x2e, 0x63, 0x6c, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x03, 0x61, 0x6c, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x48, 0x00, 0x52, 0x03, 0x61, 0x6c, 0x6c, 0x42, 0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x22, 0x51, 0x0a, 0x0b, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x4f, 0x72, 0x41, 0x6e,
	0x79, 0x12, 0x25, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0b, 0x2e, 0x63, 0x6c, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x48, 0x00,
	0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x03, 0x61, 0x6e, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x03, 0x61, 0x6e, 0x79, 0x42, 0x07, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x19, 0x0a, 0x17, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43, 0x61, 0x75, 0x73, 0x65,
	0x22, 0x36, 0x0a, 0x08, 0x4f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x78, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x74, 0x78, 0x69, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x6e, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x06, 0x6f, 0x75, 0x74, 0x6e, 0x75, 0x6d, 0x22, 0x8c, 0x01, 0x0a, 0x07, 0x46, 0x65, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x04, 0x73, 0x6c, 0x6f, 0x77, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x48, 0x00, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x77, 0x12, 0x18, 0x0a, 0x06, 0x6e, 0x6f,
	0x72, 0x6d, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x06, 0x6e, 0x6f,
	0x72, 0x6d, 0x61, 0x6c, 0x12, 0x18, 0x0a, 0x06, 0x75, 0x72, 0x67, 0x65, 0x6e, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x06, 0x75, 0x72, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x16,
	0x0a, 0x05, 0x70, 0x65, 0x72, 0x6b, 0x62, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x00, 0x52,
	0x05, 0x70, 0x65, 0x72, 0x6b, 0x62, 0x12, 0x16, 0x0a, 0x05, 0x70, 0x65, 0x72, 0x6b, 0x77, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x00, 0x52, 0x05, 0x70, 0x65, 0x72, 0x6b, 0x77, 0x42, 0x07,
	0x0a, 0x05, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x22, 0x4b, 0x0a, 0x0a, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x44, 0x65, 0x73, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x23, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0b, 0x2e, 0x63, 0x6c, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x06, 0x61, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x22, 0xa7, 0x01, 0x0a, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x48, 0x6f,
	0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x28, 0x0a, 0x10, 0x73, 0x68, 0x6f, 0x72, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x68, 0x6f,
	0x72, 0x74, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x07, 0x66,
	0x65, 0x65, 0x62, 0x61, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x63,
	0x6c, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x07, 0x66, 0x65, 0x65, 0x62, 0x61,
	0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x65, 0x65, 0x70, 0x72, 0x6f, 0x70, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x07, 0x66, 0x65, 0x65, 0x70, 0x72, 0x6f, 0x70, 0x12, 0x20, 0x0a, 0x0b,
	0x65, 0x78, 0x70, 0x69, 0x72, 0x79, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0b, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x22, 0x2e,
	0x0a, 0x09, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x68, 0x69, 0x6e, 0x74, 0x12, 0x21, 0x0a, 0x04, 0x68,
	0x6f, 0x70, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6c, 0x6e, 0x2e,
	0x52, 0x6f, 0x75, 0x74, 0x65, 0x48, 0x6f, 0x70, 0x52, 0x04, 0x68, 0x6f, 0x70, 0x73, 0x22, 0x35,
	0x0a, 0x0d, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x68, 0x69, 0x6e, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12,
	0x24, 0x0a, 0x05, 0x68, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x63, 0x6c, 0x6e, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x68, 0x69, 0x6e, 0x74, 0x52, 0x05,
	0x68, 0x69, 0x6e, 0x74, 0x73, 0x22, 0x34, 0x0a, 0x08, 0x54, 0x6c, 0x76, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x34, 0x0a, 0x09, 0x54,
	0x6c, 0x76, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x27, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6c, 0x6e, 0x2e,
	0x54, 0x6c, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x2a, 0x24, 0x0a, 0x0b, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x53, 0x69, 0x64, 0x65,
	0x12, 0x09, 0x0a, 0x05, 0x4c, 0x4f, 0x43, 0x41, 0x4c, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x52,
	0x45, 0x4d, 0x4f, 0x54, 0x45, 0x10, 0x01, 0x2a, 0x84, 0x02, 0x0a, 0x0c, 0x43, 0x68, 0x61, 0x6e,
	0x6e, 0x65, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x4f, 0x70, 0x65, 0x6e,
	0x69, 0x6e, 0x67, 0x64, 0x10, 0x00, 0x12, 0x1a, 0x0a, 0x16, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x64, 0x41, 0x77, 0x61, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x4c, 0x6f, 0x63, 0x6b, 0x69, 0x6e,
	0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x64, 0x4e, 0x6f,
	0x72, 0x6d, 0x61, 0x6c, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x64, 0x53, 0x68, 0x75, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x44, 0x6f, 0x77, 0x6e, 0x10, 0x03,
	0x12, 0x17, 0x0a, 0x13, 0x43, 0x6c, 0x6f, 0x73, 0x69, 0x6e, 0x67, 0x64, 0x53, 0x69, 0x67, 0x65,
	0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x10, 0x04, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x6c, 0x6f,
	0x73, 0x69, 0x6e, 0x67, 0x64, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x10, 0x05, 0x12,
	0x16, 0x0a, 0x12, 0x41, 0x77, 0x61, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x55, 0x6e, 0x69, 0x6c, 0x61,
	0x74, 0x65, 0x72, 0x61, 0x6c, 0x10, 0x06, 0x12, 0x14, 0x0a, 0x10, 0x46, 0x75, 0x6e, 0x64, 0x69,
	0x6e, 0x67, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x53, 0x65, 0x65, 0x6e, 0x10, 0x07, 0x12, 0x0b, 0x0a,
	0x07, 0x4f, 0x6e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x10, 0x08, 0x12, 0x15, 0x0a, 0x11, 0x44, 0x75,
	0x61, 0x6c, 0x6f, 0x70, 0x65, 0x6e, 0x64, 0x4f, 0x70, 0x65, 0x6e, 0x49, 0x6e, 0x69, 0x74, 0x10,
	0x09, 0x12, 0x1b, 0x0a, 0x17, 0x44, 0x75, 0x61, 0x6c, 0x6f, 0x70, 0x65, 0x6e, 0x64, 0x41, 0x77,
	0x61, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x4c, 0x6f, 0x63, 0x6b, 0x69, 0x6e, 0x10, 0x0a, 0x2a, 0x8a,
	0x02, 0x0a, 0x09, 0x48, 0x74, 0x6c, 0x63, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0f, 0x0a, 0x0b,
	0x53, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x48, 0x74, 0x6c, 0x63, 0x10, 0x00, 0x12, 0x11, 0x0a,
	0x0d, 0x53, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x01,
	0x12, 0x15, 0x0a, 0x11, 0x52, 0x63, 0x76, 0x64, 0x41, 0x64, 0x64, 0x52, 0x65, 0x76, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x52, 0x63, 0x76, 0x64, 0x41,
	0x64, 0x64, 0x41, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x03, 0x12, 0x18, 0x0a,
	0x14, 0x53, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x41, 0x63, 0x6b, 0x52, 0x65, 0x76, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x04, 0x12, 0x18, 0x0a, 0x14, 0x52, 0x63, 0x76, 0x64, 0x41,
	0x64, 0x64, 0x41, 0x63, 0x6b, 0x52, 0x65, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10,
	0x05, 0x12, 0x12, 0x0a, 0x0e, 0x52, 0x63, 0x76, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x48,
	0x74, 0x6c, 0x63, 0x10, 0x06, 0x12, 0x14, 0x0a, 0x10, 0x52, 0x63, 0x76, 0x64, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x07, 0x12, 0x18, 0x0a, 0x14, 0x53,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x76, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x10, 0x08, 0x12, 0x17, 0x0a, 0x13, 0x53, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x41, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x09, 0x12, 0x1b,
	0x0a, 0x17, 0x52, 0x63, 0x76, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x41, 0x63, 0x6b, 0x52,
	0x65, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x0a, 0x42, 0x1f, 0x5a, 0x1d, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x6e, 0x63, 0x61, 0x70, 0x69,
	0x74, 0x61, 0x6c, 0x2f, 0x74, 0x6f, 0x72, 0x71, 0x2f, 0x63, 0x6c, 0x6e, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_cln_primitives_proto_rawDescOnce sync.Once
	file_proto_cln_primitives_proto_rawDescData = file_proto_cln_primitives_proto_rawDesc
)

func file_proto_cln_primitives_proto_rawDescGZIP() []byte {
	file_proto_cln_primitives_proto_rawDescOnce.Do(func() {
		file_proto_cln_primitives_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_cln_primitives_proto_rawDescData)
	})
	return file_proto_cln_primitives_proto_rawDescData
}

var file_proto_cln_primitives_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_proto_cln_primitives_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_proto_cln_primitives_proto_goTypes = []interface{}{
	(ChannelSide)(0),                // 0: cln.ChannelSide
	(ChannelState)(0),               // 1: cln.ChannelState
	(HtlcState)(0),                  // 2: cln.HtlcState
	(*Amount)(nil),                  // 3: cln.Amount
	(*AmountOrAll)(nil),             // 4: cln.AmountOrAll
	(*AmountOrAny)(nil),             // 5: cln.AmountOrAny
	(*ChannelStateChangeCause)(nil), // 6: cln.ChannelStateChangeCause
	(*Outpoint)(nil),                // 7: cln.Outpoint
	(*Feerate)(nil),                 // 8: cln.Feerate
	(*OutputDesc)(nil),              // 9: cln.OutputDesc
	(*RouteHop)(nil),                // 10: cln.RouteHop
	(*Routehint)(nil),               // 11: cln.Routehint
	(*RoutehintList)(nil),           // 12: cln.RoutehintList
	(*TlvEntry)(nil),                // 13: cln.TlvEntry
	(*TlvStream)(nil),               // 14: cln.TlvStream
}
var file_proto_cln_primitives_proto_depIdxs = []int32{
	3,  // 0: cln.AmountOrAll.amount:type_name -> cln.Amount
	3,  // 1: cln.AmountOrAny.amount:type_name -> cln.Amount
	3,  // 2: cln.OutputDesc.amount:type_name -> cln.Amount
	3,  // 3: cln.RouteHop.feebase:type_name -> cln.Amount
	10, // 4: cln.Routehint.hops:type_name -> cln.RouteHop
	11, // 5: cln.RoutehintList.hints:type_name -> cln.Routehint
	13, // 6: cln.TlvStream.entries:type_name -> cln.TlvEntry
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_proto_cln_primitives_proto_init() }
func file_proto_cln_primitives_proto_init() {
	if File_proto_cln_primitives_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_cln_primitives_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Amount); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AmountOrAll); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AmountOrAny); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChannelStateChangeCause); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Outpoint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feerate); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OutputDesc); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RouteHop); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Routehint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RoutehintList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TlvEntry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_cln_primitives_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TlvStream); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_proto_cln_primitives_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*AmountOrAll_Amount)(nil),
		(*AmountOrAll_All)(nil),
	}
	file_proto_cln_primitives_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*AmountOrAny_Amount)(nil),
		(*AmountOrAny_Any)(nil),
	}
	file_proto_cln_primitives_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*Feerate_Slow)(nil),
		(*Feerate_Normal)(nil),
		(*Feerate_Urgent)(nil),
		(*Feerate_Perkb)(nil),
		(*Feerate_Perkw)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_cln_primitives_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_cln_primitives_proto_goTypes,
		DependencyIndexes: file_proto_cln_primitives_proto_depIdxs,
		EnumInfos:         file_proto_cln_primitives_proto_enumTypes,
		MessageInfos:      file_proto_cln_primitives_proto_msgTypes,
	}.Build()
	File_proto_cln_primitives_proto = out.File
	file_proto_cln_primitives_proto_rawDesc = nil
	file_proto_cln_primitives_proto_goTypes = nil
	file_proto_cln_primitives_proto_depIdxs = nil
}
