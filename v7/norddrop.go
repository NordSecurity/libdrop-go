
package norddrop

// #include <norddrop.h>
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	"encoding/binary"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)



type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer {
		capacity: C.int(b.Capacity()),
		len: C.int(b.Len()),
		data: (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.ffi_norddrop_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}


func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes {
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}
	
	return rustCall(func( status *C.RustCallStatus) RustBuffer {
		return C.ffi_norddrop_rustbuffer_from_bytes(foreign, status)
	})
}



type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}



func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}


func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}


func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {
        
        (&FfiConverterCallbackInterfaceEventCallback{}).register();
        (&FfiConverterCallbackInterfaceFdResolver{}).register();
        (&FfiConverterCallbackInterfaceKeyStore{}).register();
        (&FfiConverterCallbackInterfaceLogger{}).register();
        uniffiCheckChecksums()
}


func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_norddrop_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: UniFFI contract version mismatch")
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_func_version(uniffiStatus)
	})
	if checksum != 54120 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_func_version: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_download_file(uniffiStatus)
	})
	if checksum != 32954 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_download_file: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_finalize_transfer(uniffiStatus)
	})
	if checksum != 37465 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_finalize_transfer: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_network_refresh(uniffiStatus)
	})
	if checksum != 53685 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_network_refresh: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_new_transfer(uniffiStatus)
	})
	if checksum != 51452 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_new_transfer: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_purge_transfers(uniffiStatus)
	})
	if checksum != 55603 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_purge_transfers: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_purge_transfers_until(uniffiStatus)
	})
	if checksum != 13823 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_purge_transfers_until: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_reject_file(uniffiStatus)
	})
	if checksum != 19601 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_reject_file: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_remove_file(uniffiStatus)
	})
	if checksum != 38631 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_remove_file: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_set_fd_resolver(uniffiStatus)
	})
	if checksum != 35835 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_set_fd_resolver: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_start(uniffiStatus)
	})
	if checksum != 57097 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_start: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_stop(uniffiStatus)
	})
	if checksum != 43969 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_stop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_norddrop_transfers_since(uniffiStatus)
	})
	if checksum != 16492 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_norddrop_transfers_since: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_constructor_norddrop_new(uniffiStatus)
	})
	if checksum != 18531 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_constructor_norddrop_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_eventcallback_on_event(uniffiStatus)
	})
	if checksum != 57627 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_eventcallback_on_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_fdresolver_on_fd(uniffiStatus)
	})
	if checksum != 41805 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_fdresolver_on_fd: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_keystore_on_pubkey(uniffiStatus)
	})
	if checksum != 31192 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_keystore_on_pubkey: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_keystore_privkey(uniffiStatus)
	})
	if checksum != 24684 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_keystore_privkey: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_logger_on_log(uniffiStatus)
	})
	if checksum != 26642 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_logger_on_log: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_norddrop_checksum_method_logger_level(uniffiStatus)
	})
	if checksum != 56965 {
		// If this happens try cleaning and rebuilding your project
		panic("norddrop: uniffi_norddrop_checksum_method_logger_level: UniFFI API checksum mismatch")
	}
	}
}




type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct {}

func (FfiDestroyerUint32) Destroy(_ uint32) {}


type FfiConverterInt32 struct{}

var FfiConverterInt32INSTANCE = FfiConverterInt32{}

func (FfiConverterInt32) Lower(value int32) C.int32_t {
	return C.int32_t(value)
}

func (FfiConverterInt32) Write(writer io.Writer, value int32) {
	writeInt32(writer, value)
}

func (FfiConverterInt32) Lift(value C.int32_t) int32 {
	return int32(value)
}

func (FfiConverterInt32) Read(reader io.Reader) int32 {
	return readInt32(reader)
}

type FfiDestroyerInt32 struct {}

func (FfiDestroyerInt32) Destroy(_ int32) {}


type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct {}

func (FfiDestroyerUint64) Destroy(_ uint64) {}


type FfiConverterInt64 struct{}

var FfiConverterInt64INSTANCE = FfiConverterInt64{}

func (FfiConverterInt64) Lower(value int64) C.int64_t {
	return C.int64_t(value)
}

func (FfiConverterInt64) Write(writer io.Writer, value int64) {
	writeInt64(writer, value)
}

func (FfiConverterInt64) Lift(value C.int64_t) int64 {
	return int64(value)
}

func (FfiConverterInt64) Read(reader io.Reader) int64 {
	return readInt64(reader)
}

type FfiDestroyerInt64 struct {}

func (FfiDestroyerInt64) Destroy(_ int64) {}


type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct {}

func (FfiDestroyerBool) Destroy(_ bool) {}


type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct {}

func (FfiDestroyerString) Destroy(_ string) {}


type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) Lower(value []byte) RustBuffer {
	return LowerIntoRustBuffer[[]byte](c, value)
}

func (c FfiConverterBytes) Write(writer io.Writer, value []byte) {
	if len(value) > math.MaxInt32 {
		panic("[]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := writer.Write(value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing []byte, expected %d, written %d", len(value), write_length))
	}
}

func (c FfiConverterBytes) Lift(rb RustBufferI) []byte {
	return LiftFromRustBuffer[[]byte](c, rb)
}

func (c FfiConverterBytes) Read(reader io.Reader) []byte {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct {}

func (FfiDestroyerBytes) Destroy(_ []byte) {}




// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer unsafe.Pointer
	callCounter atomic.Int64
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject {
		pointer: pointer,
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject)incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter + 1) {
			break
		}
	}

	return ffiObject.pointer
}

func (ffiObject *FfiObject)decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject)destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject)freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}
type NordDrop struct {
	ffiObject FfiObject
}
// Create a new instance of norddrop. This is a required step to work
// with API further
//
// # Arguments
// * `event_cb` - Event callback
// * `logger` - Logger callback
// * `key_store` - Fetches peer's public key and provides own private key. 
func NewNordDrop(eventCb EventCallback, keyStore KeyStore, logger Logger) (*NordDrop, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_norddrop_fn_constructor_norddrop_new(FfiConverterCallbackInterfaceEventCallbackINSTANCE.Lower(eventCb), FfiConverterCallbackInterfaceKeyStoreINSTANCE.Lower(keyStore), FfiConverterCallbackInterfaceLoggerINSTANCE.Lower(logger), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *NordDrop
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterNordDropINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}




// # Download a file from the peer
//
// # Arguments
// * `transfer_id` - Transfer UUID
// * `file_id` - File ID
// * `destination` - Destination path
func (_self *NordDrop)DownloadFile(transferId string, fileId string, destination string) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_download_file(
		_pointer,FfiConverterStringINSTANCE.Lower(transferId), FfiConverterStringINSTANCE.Lower(fileId), FfiConverterStringINSTANCE.Lower(destination), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// # Finalizes the transfer from either side
//
// # Arguments
// * `transfer_id`: Transfer UUID
func (_self *NordDrop)FinalizeTransfer(transferId string) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_finalize_transfer(
		_pointer,FfiConverterStringINSTANCE.Lower(transferId), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Refresh connections. Should be called when anything about the network
// changes that might affect connections. Also when peer availability has
// changed. This will kick-start the automated retries for all transfers.
func (_self *NordDrop)NetworkRefresh() error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_network_refresh(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Initialize a new transfer with the provided peer and descriptors
//
// # Arguments
// * `peer` - Peer address.
// * `descriptors` - transfer file descriptors.
//
// # Returns
// A String containing the transfer UUID.
func (_self *NordDrop)NewTransfer(peer string, descriptors []TransferDescriptor) (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_norddrop_fn_method_norddrop_new_transfer(
		_pointer,FfiConverterStringINSTANCE.Lower(peer), FfiConverterSequenceTypeTransferDescriptorINSTANCE.Lower(descriptors), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


// Purge transfers from the database
//
// # Arguments
// * `transfer_ids` - array of transfer UUIDs
func (_self *NordDrop)PurgeTransfers(transferIds []string) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_purge_transfers(
		_pointer,FfiConverterSequenceStringINSTANCE.Lower(transferIds), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Purge transfers from the database until the given timestamp
//
// # Arguments
// * `until` - Unix timestamp in milliseconds
func (_self *NordDrop)PurgeTransfersUntil(until int64) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_purge_transfers_until(
		_pointer,FfiConverterInt64INSTANCE.Lower(until), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Reject a file from either side
//
// # Arguments
// * `transfer_id`: Transfer UUID
// * `file_id`: File ID
func (_self *NordDrop)RejectFile(transferId string, fileId string) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_reject_file(
		_pointer,FfiConverterStringINSTANCE.Lower(transferId), FfiConverterStringINSTANCE.Lower(fileId), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Removes a single transfer file from the database. The file must be in
// the **terminal** state beforehand, otherwise the error is returned.
//
//  # Arguments
// * `transfer_id`: Transfer UUID
// * `file_id`: File ID
func (_self *NordDrop)RemoveFile(transferId string, fileId string) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_remove_file(
		_pointer,FfiConverterStringINSTANCE.Lower(transferId), FfiConverterStringINSTANCE.Lower(fileId), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Set a file descriptor (FD) resolver callback.
// The callback provides FDs based on URI.
// This function should be called before `start()`, otherwise it will
// return an error.
//
// # Arguments
// * `resolver`: The resolver structure
//
// # Warning
// This function is intended to be called only on UNIX platforms
func (_self *NordDrop)SetFdResolver(resolver FdResolver) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_set_fd_resolver(
		_pointer,FfiConverterCallbackInterfaceFdResolverINSTANCE.Lower(resolver), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Starts libdrop
//
// # Arguments
// * `addr` - Address to listen on
// * `config` - configuration
//
// # Configuration Parameters
//
// * `dir_depth_limit` - if the tree contains more levels then the error is
// returned.
//
// * `transfer_file_limit` - when aggregating files from the path, if this
// limit is reached, an error is returned.
//
// * `moose_event_path` - moose database path.
//
// * `moose_prod` - moose production flag.
//
// * `storage_path` - storage path for persistence engine.
//
// * `checksum_events_size_threshold_bytes` - emit checksum events only if file
//   is equal or greater than this size. If omited, no checksumming events are
//   emited.
//
// # Safety
// The pointers provided must be valid
func (_self *NordDrop)Start(addr string, config Config) error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_start(
		_pointer,FfiConverterStringINSTANCE.Lower(addr), FfiConverterTypeConfigINSTANCE.Lower(config), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Stop norddrop instance
func (_self *NordDrop)Stop() error {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_norddrop_fn_method_norddrop_stop(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Get transfers from the database
//
// # Arguments
// * `since_timestamp` - UNIX timestamp in milliseconds
func (_self *NordDrop)TransfersSince(since int64) ([]TransferInfo, error) {
	_pointer := _self.ffiObject.incrementPointer("*NordDrop")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeLibdropError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_norddrop_fn_method_norddrop_transfers_since(
		_pointer,FfiConverterInt64INSTANCE.Lower(since), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue []TransferInfo
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterSequenceTypeTransferInfoINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



func (object *NordDrop)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNordDrop struct {}

var FfiConverterNordDropINSTANCE = FfiConverterNordDrop{}

func (c FfiConverterNordDrop) Lift(pointer unsafe.Pointer) *NordDrop {
	result := &NordDrop {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_norddrop_fn_free_norddrop(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, (*NordDrop).Destroy)
	return result
}

func (c FfiConverterNordDrop) Read(reader io.Reader) *NordDrop {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNordDrop) Lower(value *NordDrop) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*NordDrop")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNordDrop) Write(writer io.Writer, value *NordDrop) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNordDrop struct {}

func (_ FfiDestroyerNordDrop) Destroy(value *NordDrop) {
	value.Destroy()
}


// The configuration structure
type Config struct {
	// If the transfer directory tree contains more levels then the error is
	// returned.
	DirDepthLimit uint64
	// When aggregating files from the path, if this limit is reached, an error
	// is returned.
	TransferFileLimit uint64
	// Moose database path
	MooseEventPath string
	// Moose production flag
	MooseProd bool
	// Storage path for persistence engine
	StoragePath string
	// Emit checksum events only if file is equal or greater than this size. 
	// If omited, no checksumming events are emited.
	ChecksumEventsSizeThreshold *uint64
	// Emit checksum events at set granularity
	ChecksumEventsGranularity *uint64
	// Limits the number of connection retries afer the `network_refresh()` call.
	ConnectionRetries *uint32
}

func (r *Config) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.DirDepthLimit);
		FfiDestroyerUint64{}.Destroy(r.TransferFileLimit);
		FfiDestroyerString{}.Destroy(r.MooseEventPath);
		FfiDestroyerBool{}.Destroy(r.MooseProd);
		FfiDestroyerString{}.Destroy(r.StoragePath);
		FfiDestroyerOptionalUint64{}.Destroy(r.ChecksumEventsSizeThreshold);
		FfiDestroyerOptionalUint64{}.Destroy(r.ChecksumEventsGranularity);
		FfiDestroyerOptionalUint32{}.Destroy(r.ConnectionRetries);
}

type FfiConverterTypeConfig struct {}

var FfiConverterTypeConfigINSTANCE = FfiConverterTypeConfig{}

func (c FfiConverterTypeConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterTypeConfig) Read(reader io.Reader) Config {
	return Config {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeConfig) Lower(value Config) RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterTypeConfig) Write(writer io.Writer, value Config) {
		FfiConverterUint64INSTANCE.Write(writer, value.DirDepthLimit);
		FfiConverterUint64INSTANCE.Write(writer, value.TransferFileLimit);
		FfiConverterStringINSTANCE.Write(writer, value.MooseEventPath);
		FfiConverterBoolINSTANCE.Write(writer, value.MooseProd);
		FfiConverterStringINSTANCE.Write(writer, value.StoragePath);
		FfiConverterOptionalUint64INSTANCE.Write(writer, value.ChecksumEventsSizeThreshold);
		FfiConverterOptionalUint64INSTANCE.Write(writer, value.ChecksumEventsGranularity);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.ConnectionRetries);
}

type FfiDestroyerTypeConfig struct {}

func (_ FfiDestroyerTypeConfig) Destroy(value Config) {
	value.Destroy()
}


// The event type emited by the library
type Event struct {
	// Creation timestamp
	Timestamp int64
	// A type of event
	Kind EventKind
}

func (r *Event) Destroy() {
		FfiDestroyerInt64{}.Destroy(r.Timestamp);
		FfiDestroyerTypeEventKind{}.Destroy(r.Kind);
}

type FfiConverterTypeEvent struct {}

var FfiConverterTypeEventINSTANCE = FfiConverterTypeEvent{}

func (c FfiConverterTypeEvent) Lift(rb RustBufferI) Event {
	return LiftFromRustBuffer[Event](c, rb)
}

func (c FfiConverterTypeEvent) Read(reader io.Reader) Event {
	return Event {
			FfiConverterInt64INSTANCE.Read(reader),
			FfiConverterTypeEventKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEvent) Lower(value Event) RustBuffer {
	return LowerIntoRustBuffer[Event](c, value)
}

func (c FfiConverterTypeEvent) Write(writer io.Writer, value Event) {
		FfiConverterInt64INSTANCE.Write(writer, value.Timestamp);
		FfiConverterTypeEventKindINSTANCE.Write(writer, value.Kind);
}

type FfiDestroyerTypeEvent struct {}

func (_ FfiDestroyerTypeEvent) Destroy(value Event) {
	value.Destroy()
}


// The description and history of a signle incoming file
type IncomingPath struct {
	// File ID
	FileId string
	// File path relative to the transfer's root directory
	RelativePath string
	// File size
	Bytes uint64
	// Curently received file bytes
	BytesReceived uint64
	// History of the file state chagnes
	States []IncomingPathState
}

func (r *IncomingPath) Destroy() {
		FfiDestroyerString{}.Destroy(r.FileId);
		FfiDestroyerString{}.Destroy(r.RelativePath);
		FfiDestroyerUint64{}.Destroy(r.Bytes);
		FfiDestroyerUint64{}.Destroy(r.BytesReceived);
		FfiDestroyerSequenceTypeIncomingPathState{}.Destroy(r.States);
}

type FfiConverterTypeIncomingPath struct {}

var FfiConverterTypeIncomingPathINSTANCE = FfiConverterTypeIncomingPath{}

func (c FfiConverterTypeIncomingPath) Lift(rb RustBufferI) IncomingPath {
	return LiftFromRustBuffer[IncomingPath](c, rb)
}

func (c FfiConverterTypeIncomingPath) Read(reader io.Reader) IncomingPath {
	return IncomingPath {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterSequenceTypeIncomingPathStateINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeIncomingPath) Lower(value IncomingPath) RustBuffer {
	return LowerIntoRustBuffer[IncomingPath](c, value)
}

func (c FfiConverterTypeIncomingPath) Write(writer io.Writer, value IncomingPath) {
		FfiConverterStringINSTANCE.Write(writer, value.FileId);
		FfiConverterStringINSTANCE.Write(writer, value.RelativePath);
		FfiConverterUint64INSTANCE.Write(writer, value.Bytes);
		FfiConverterUint64INSTANCE.Write(writer, value.BytesReceived);
		FfiConverterSequenceTypeIncomingPathStateINSTANCE.Write(writer, value.States);
}

type FfiDestroyerTypeIncomingPath struct {}

func (_ FfiDestroyerTypeIncomingPath) Destroy(value IncomingPath) {
	value.Destroy()
}


// A single change in the incoming file state
type IncomingPathState struct {
	// The creation time as a UNIX timestamp in milliseconds.
	CreatedAt int64
	// The type of the state change.
	Kind IncomingPathStateKind
}

func (r *IncomingPathState) Destroy() {
		FfiDestroyerInt64{}.Destroy(r.CreatedAt);
		FfiDestroyerTypeIncomingPathStateKind{}.Destroy(r.Kind);
}

type FfiConverterTypeIncomingPathState struct {}

var FfiConverterTypeIncomingPathStateINSTANCE = FfiConverterTypeIncomingPathState{}

func (c FfiConverterTypeIncomingPathState) Lift(rb RustBufferI) IncomingPathState {
	return LiftFromRustBuffer[IncomingPathState](c, rb)
}

func (c FfiConverterTypeIncomingPathState) Read(reader io.Reader) IncomingPathState {
	return IncomingPathState {
			FfiConverterInt64INSTANCE.Read(reader),
			FfiConverterTypeIncomingPathStateKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeIncomingPathState) Lower(value IncomingPathState) RustBuffer {
	return LowerIntoRustBuffer[IncomingPathState](c, value)
}

func (c FfiConverterTypeIncomingPathState) Write(writer io.Writer, value IncomingPathState) {
		FfiConverterInt64INSTANCE.Write(writer, value.CreatedAt);
		FfiConverterTypeIncomingPathStateKindINSTANCE.Write(writer, value.Kind);
}

type FfiDestroyerTypeIncomingPathState struct {}

func (_ FfiDestroyerTypeIncomingPathState) Destroy(value IncomingPathState) {
	value.Destroy()
}


// The description and history of a signle outgoing file
type OutgoingPath struct {
	// File ID
	FileId string
	// File path relative to the transfer's root directory
	RelativePath string
	// File size
	Bytes uint64
	// Curently transferred file bytes
	BytesSent uint64
	// The source of the file data
	Source OutgoingFileSource
	// History of the file state chagnes
	States []OutgoingPathState
}

func (r *OutgoingPath) Destroy() {
		FfiDestroyerString{}.Destroy(r.FileId);
		FfiDestroyerString{}.Destroy(r.RelativePath);
		FfiDestroyerUint64{}.Destroy(r.Bytes);
		FfiDestroyerUint64{}.Destroy(r.BytesSent);
		FfiDestroyerTypeOutgoingFileSource{}.Destroy(r.Source);
		FfiDestroyerSequenceTypeOutgoingPathState{}.Destroy(r.States);
}

type FfiConverterTypeOutgoingPath struct {}

var FfiConverterTypeOutgoingPathINSTANCE = FfiConverterTypeOutgoingPath{}

func (c FfiConverterTypeOutgoingPath) Lift(rb RustBufferI) OutgoingPath {
	return LiftFromRustBuffer[OutgoingPath](c, rb)
}

func (c FfiConverterTypeOutgoingPath) Read(reader io.Reader) OutgoingPath {
	return OutgoingPath {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterTypeOutgoingFileSourceINSTANCE.Read(reader),
			FfiConverterSequenceTypeOutgoingPathStateINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOutgoingPath) Lower(value OutgoingPath) RustBuffer {
	return LowerIntoRustBuffer[OutgoingPath](c, value)
}

func (c FfiConverterTypeOutgoingPath) Write(writer io.Writer, value OutgoingPath) {
		FfiConverterStringINSTANCE.Write(writer, value.FileId);
		FfiConverterStringINSTANCE.Write(writer, value.RelativePath);
		FfiConverterUint64INSTANCE.Write(writer, value.Bytes);
		FfiConverterUint64INSTANCE.Write(writer, value.BytesSent);
		FfiConverterTypeOutgoingFileSourceINSTANCE.Write(writer, value.Source);
		FfiConverterSequenceTypeOutgoingPathStateINSTANCE.Write(writer, value.States);
}

type FfiDestroyerTypeOutgoingPath struct {}

func (_ FfiDestroyerTypeOutgoingPath) Destroy(value OutgoingPath) {
	value.Destroy()
}


// The description and history of a signle outgoing file
type OutgoingPathState struct {
	// The creation time as a UNIX timestamp in milliseconds.
	CreatedAt int64
	// The type of the state change.
	Kind OutgoingPathStateKind
}

func (r *OutgoingPathState) Destroy() {
		FfiDestroyerInt64{}.Destroy(r.CreatedAt);
		FfiDestroyerTypeOutgoingPathStateKind{}.Destroy(r.Kind);
}

type FfiConverterTypeOutgoingPathState struct {}

var FfiConverterTypeOutgoingPathStateINSTANCE = FfiConverterTypeOutgoingPathState{}

func (c FfiConverterTypeOutgoingPathState) Lift(rb RustBufferI) OutgoingPathState {
	return LiftFromRustBuffer[OutgoingPathState](c, rb)
}

func (c FfiConverterTypeOutgoingPathState) Read(reader io.Reader) OutgoingPathState {
	return OutgoingPathState {
			FfiConverterInt64INSTANCE.Read(reader),
			FfiConverterTypeOutgoingPathStateKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOutgoingPathState) Lower(value OutgoingPathState) RustBuffer {
	return LowerIntoRustBuffer[OutgoingPathState](c, value)
}

func (c FfiConverterTypeOutgoingPathState) Write(writer io.Writer, value OutgoingPathState) {
		FfiConverterInt64INSTANCE.Write(writer, value.CreatedAt);
		FfiConverterTypeOutgoingPathStateKindINSTANCE.Write(writer, value.Kind);
}

type FfiDestroyerTypeOutgoingPathState struct {}

func (_ FfiDestroyerTypeOutgoingPathState) Destroy(value OutgoingPathState) {
	value.Destroy()
}


// The outgoing transfer file structure
type QueuedFile struct {
	// File ID
	Id string
	// File path
	Path string
	// File size
	Size uint64
	// File base directory
	BaseDir *string
}

func (r *QueuedFile) Destroy() {
		FfiDestroyerString{}.Destroy(r.Id);
		FfiDestroyerString{}.Destroy(r.Path);
		FfiDestroyerUint64{}.Destroy(r.Size);
		FfiDestroyerOptionalString{}.Destroy(r.BaseDir);
}

type FfiConverterTypeQueuedFile struct {}

var FfiConverterTypeQueuedFileINSTANCE = FfiConverterTypeQueuedFile{}

func (c FfiConverterTypeQueuedFile) Lift(rb RustBufferI) QueuedFile {
	return LiftFromRustBuffer[QueuedFile](c, rb)
}

func (c FfiConverterTypeQueuedFile) Read(reader io.Reader) QueuedFile {
	return QueuedFile {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeQueuedFile) Lower(value QueuedFile) RustBuffer {
	return LowerIntoRustBuffer[QueuedFile](c, value)
}

func (c FfiConverterTypeQueuedFile) Write(writer io.Writer, value QueuedFile) {
		FfiConverterStringINSTANCE.Write(writer, value.Id);
		FfiConverterStringINSTANCE.Write(writer, value.Path);
		FfiConverterUint64INSTANCE.Write(writer, value.Size);
		FfiConverterOptionalStringINSTANCE.Write(writer, value.BaseDir);
}

type FfiDestroyerTypeQueuedFile struct {}

func (_ FfiDestroyerTypeQueuedFile) Destroy(value QueuedFile) {
	value.Destroy()
}


// The incoming transfer file structure
type ReceivedFile struct {
	// File ID
	Id string
	// File path
	Path string
	// File size
	Size uint64
}

func (r *ReceivedFile) Destroy() {
		FfiDestroyerString{}.Destroy(r.Id);
		FfiDestroyerString{}.Destroy(r.Path);
		FfiDestroyerUint64{}.Destroy(r.Size);
}

type FfiConverterTypeReceivedFile struct {}

var FfiConverterTypeReceivedFileINSTANCE = FfiConverterTypeReceivedFile{}

func (c FfiConverterTypeReceivedFile) Lift(rb RustBufferI) ReceivedFile {
	return LiftFromRustBuffer[ReceivedFile](c, rb)
}

func (c FfiConverterTypeReceivedFile) Read(reader io.Reader) ReceivedFile {
	return ReceivedFile {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeReceivedFile) Lower(value ReceivedFile) RustBuffer {
	return LowerIntoRustBuffer[ReceivedFile](c, value)
}

func (c FfiConverterTypeReceivedFile) Write(writer io.Writer, value ReceivedFile) {
		FfiConverterStringINSTANCE.Write(writer, value.Id);
		FfiConverterStringINSTANCE.Write(writer, value.Path);
		FfiConverterUint64INSTANCE.Write(writer, value.Size);
}

type FfiDestroyerTypeReceivedFile struct {}

func (_ FfiDestroyerTypeReceivedFile) Destroy(value ReceivedFile) {
	value.Destroy()
}


// The common state structure
type Status struct {
	// Status code
	Status StatusCode
	// OS error number if available
	OsErrorCode *int32
}

func (r *Status) Destroy() {
		FfiDestroyerTypeStatusCode{}.Destroy(r.Status);
		FfiDestroyerOptionalInt32{}.Destroy(r.OsErrorCode);
}

type FfiConverterTypeStatus struct {}

var FfiConverterTypeStatusINSTANCE = FfiConverterTypeStatus{}

func (c FfiConverterTypeStatus) Lift(rb RustBufferI) Status {
	return LiftFromRustBuffer[Status](c, rb)
}

func (c FfiConverterTypeStatus) Read(reader io.Reader) Status {
	return Status {
			FfiConverterTypeStatusCodeINSTANCE.Read(reader),
			FfiConverterOptionalInt32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeStatus) Lower(value Status) RustBuffer {
	return LowerIntoRustBuffer[Status](c, value)
}

func (c FfiConverterTypeStatus) Write(writer io.Writer, value Status) {
		FfiConverterTypeStatusCodeINSTANCE.Write(writer, value.Status);
		FfiConverterOptionalInt32INSTANCE.Write(writer, value.OsErrorCode);
}

type FfiDestroyerTypeStatus struct {}

func (_ FfiDestroyerTypeStatus) Destroy(value Status) {
	value.Destroy()
}


// Transfer and files in it contain history of states that can be used to
// replay what happened and the last state denotes the current state of the
// transfer.
type TransferInfo struct {
	// Transfer UUID
	Id string
	// The creation time as a UNIX timestamp in milliseconds.
	CreatedAt int64
	// Peer's IP address
	Peer string
	// History of transfer states
	States []TransferState
	// The transfer type description
	Kind TransferKind
}

func (r *TransferInfo) Destroy() {
		FfiDestroyerString{}.Destroy(r.Id);
		FfiDestroyerInt64{}.Destroy(r.CreatedAt);
		FfiDestroyerString{}.Destroy(r.Peer);
		FfiDestroyerSequenceTypeTransferState{}.Destroy(r.States);
		FfiDestroyerTypeTransferKind{}.Destroy(r.Kind);
}

type FfiConverterTypeTransferInfo struct {}

var FfiConverterTypeTransferInfoINSTANCE = FfiConverterTypeTransferInfo{}

func (c FfiConverterTypeTransferInfo) Lift(rb RustBufferI) TransferInfo {
	return LiftFromRustBuffer[TransferInfo](c, rb)
}

func (c FfiConverterTypeTransferInfo) Read(reader io.Reader) TransferInfo {
	return TransferInfo {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterInt64INSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterSequenceTypeTransferStateINSTANCE.Read(reader),
			FfiConverterTypeTransferKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTransferInfo) Lower(value TransferInfo) RustBuffer {
	return LowerIntoRustBuffer[TransferInfo](c, value)
}

func (c FfiConverterTypeTransferInfo) Write(writer io.Writer, value TransferInfo) {
		FfiConverterStringINSTANCE.Write(writer, value.Id);
		FfiConverterInt64INSTANCE.Write(writer, value.CreatedAt);
		FfiConverterStringINSTANCE.Write(writer, value.Peer);
		FfiConverterSequenceTypeTransferStateINSTANCE.Write(writer, value.States);
		FfiConverterTypeTransferKindINSTANCE.Write(writer, value.Kind);
}

type FfiDestroyerTypeTransferInfo struct {}

func (_ FfiDestroyerTypeTransferInfo) Destroy(value TransferInfo) {
	value.Destroy()
}


// A single change in the transfer state
type TransferState struct {
	// The creation time as a UNIX timestamp in milliseconds.
	CreatedAt int64
	// The type of the state change.
	Kind TransferStateKind
}

func (r *TransferState) Destroy() {
		FfiDestroyerInt64{}.Destroy(r.CreatedAt);
		FfiDestroyerTypeTransferStateKind{}.Destroy(r.Kind);
}

type FfiConverterTypeTransferState struct {}

var FfiConverterTypeTransferStateINSTANCE = FfiConverterTypeTransferState{}

func (c FfiConverterTypeTransferState) Lift(rb RustBufferI) TransferState {
	return LiftFromRustBuffer[TransferState](c, rb)
}

func (c FfiConverterTypeTransferState) Read(reader io.Reader) TransferState {
	return TransferState {
			FfiConverterInt64INSTANCE.Read(reader),
			FfiConverterTypeTransferStateKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTransferState) Lower(value TransferState) RustBuffer {
	return LowerIntoRustBuffer[TransferState](c, value)
}

func (c FfiConverterTypeTransferState) Write(writer io.Writer, value TransferState) {
		FfiConverterInt64INSTANCE.Write(writer, value.CreatedAt);
		FfiConverterTypeTransferStateKindINSTANCE.Write(writer, value.Kind);
}

type FfiDestroyerTypeTransferState struct {}

func (_ FfiDestroyerTypeTransferState) Destroy(value TransferState) {
	value.Destroy()
}



// Possible types of events
type EventKind interface {
	Destroy()
}
// Emitted when the application receives a transfer request from the peer. It
// contains the peer IP address, transfer ID, and file list.
type EventKindRequestReceived struct {
	Peer string
	TransferId string
	Files []ReceivedFile
}

func (e EventKindRequestReceived) Destroy() {
		FfiDestroyerString{}.Destroy(e.Peer);
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerSequenceTypeReceivedFile{}.Destroy(e.Files);
}
// Emitted when the application creates a transfer.
type EventKindRequestQueued struct {
	Peer string
	TransferId string
	Files []QueuedFile
}

func (e EventKindRequestQueued) Destroy() {
		FfiDestroyerString{}.Destroy(e.Peer);
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerSequenceTypeQueuedFile{}.Destroy(e.Files);
}
// Emitted when a file transfer is started. Valid for both sending and
// receiving peers.
type EventKindFileStarted struct {
	TransferId string
	FileId string
	Transferred uint64
}

func (e EventKindFileStarted) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.Transferred);
}
// Emitted whenever an amount of data for a single file is transferred between
// peers. Valid for both sending and receiving peers.
type EventKindFileProgress struct {
	TransferId string
	FileId string
	Transferred uint64
}

func (e EventKindFileProgress) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.Transferred);
}
// The file has been successfully downloaded.
type EventKindFileDownloaded struct {
	TransferId string
	FileId string
	FinalPath string
}

func (e EventKindFileDownloaded) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerString{}.Destroy(e.FinalPath);
}
// The file has been successfully uploaded.
type EventKindFileUploaded struct {
	TransferId string
	FileId string
}

func (e EventKindFileUploaded) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
}
// File transfer has failed.
type EventKindFileFailed struct {
	TransferId string
	FileId string
	Status Status
}

func (e EventKindFileFailed) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerTypeStatus{}.Destroy(e.Status);
}
// The file was rejected.
type EventKindFileRejected struct {
	TransferId string
	FileId string
	ByPeer bool
}

func (e EventKindFileRejected) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerBool{}.Destroy(e.ByPeer);
}
// Emited automatically for each file in flight in case the peer goes offline
// but the transfer will be resumed.
type EventKindFilePaused struct {
	TransferId string
	FileId string
}

func (e EventKindFilePaused) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
}
// The event may be emitted before the outgoing file is started. It’s an indication
// of a delayed transfer because of too many active outgoing files in flight.
// Whenever the number of active files decreases the file will proceed with the
// TransferStarted event. Valid for sending peers.
type EventKindFileThrottled struct {
	TransferId string
	FileId string
	Transferred uint64
}

func (e EventKindFileThrottled) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.Transferred);
}
// Indicates that the file transfer is registered and ready. It is emitted as a
// response to the `download()` call.
type EventKindFilePending struct {
	TransferId string
	FileId string
}

func (e EventKindFilePending) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
}
// Transfer is finalized and no further action on the transfer are possible.
type EventKindTransferFinalized struct {
	TransferId string
	ByPeer bool
}

func (e EventKindTransferFinalized) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerBool{}.Destroy(e.ByPeer);
}
// The whole transfer has failed.
type EventKindTransferFailed struct {
	TransferId string
	Status Status
}

func (e EventKindTransferFailed) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerTypeStatus{}.Destroy(e.Status);
}
// Indicates that the connection made towards the peer was unsuccessful. It might
// be emitted as a response to the `network_refresh()` call.
type EventKindTransferDeferred struct {
	TransferId string
	Peer string
	Status Status
}

func (e EventKindTransferDeferred) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.Peer);
		FfiDestroyerTypeStatus{}.Destroy(e.Status);
}
// On the downloader side is emitted when the checksum calculation starts. It
// happens after the download.
type EventKindFinalizeChecksumStarted struct {
	TransferId string
	FileId string
	Size uint64
}

func (e EventKindFinalizeChecksumStarted) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.Size);
}
// Reports finalize checksum finished(downloader side only).
type EventKindFinalizeChecksumFinished struct {
	TransferId string
	FileId string
}

func (e EventKindFinalizeChecksumFinished) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
}
// Reports finalize checksumming progress(downloader side only).
type EventKindFinalizeChecksumProgress struct {
	TransferId string
	FileId string
	BytesChecksummed uint64
}

func (e EventKindFinalizeChecksumProgress) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.BytesChecksummed);
}
// On the downloader side is emitted when the checksum calculation starts. It
// happens when resuming the download.
type EventKindVerifyChecksumStarted struct {
	TransferId string
	FileId string
	Size uint64
}

func (e EventKindVerifyChecksumStarted) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.Size);
}
// Reports verify checksum finished(downloader side only).
type EventKindVerifyChecksumFinished struct {
	TransferId string
	FileId string
}

func (e EventKindVerifyChecksumFinished) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
}
// Reports verify checksumming progress(downloader side only).
type EventKindVerifyChecksumProgress struct {
	TransferId string
	FileId string
	BytesChecksummed uint64
}

func (e EventKindVerifyChecksumProgress) Destroy() {
		FfiDestroyerString{}.Destroy(e.TransferId);
		FfiDestroyerString{}.Destroy(e.FileId);
		FfiDestroyerUint64{}.Destroy(e.BytesChecksummed);
}
// This event is used to indicate some runtime error that is not related to the
// transfer. For example database errors due to automatic retries.
type EventKindRuntimeError struct {
	Status StatusCode
}

func (e EventKindRuntimeError) Destroy() {
		FfiDestroyerTypeStatusCode{}.Destroy(e.Status);
}

type FfiConverterTypeEventKind struct {}

var FfiConverterTypeEventKindINSTANCE = FfiConverterTypeEventKind{}

func (c FfiConverterTypeEventKind) Lift(rb RustBufferI) EventKind {
	return LiftFromRustBuffer[EventKind](c, rb)
}

func (c FfiConverterTypeEventKind) Lower(value EventKind) RustBuffer {
	return LowerIntoRustBuffer[EventKind](c, value)
}
func (FfiConverterTypeEventKind) Read(reader io.Reader) EventKind {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return EventKindRequestReceived{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterSequenceTypeReceivedFileINSTANCE.Read(reader),
			};
		case 2:
			return EventKindRequestQueued{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterSequenceTypeQueuedFileINSTANCE.Read(reader),
			};
		case 3:
			return EventKindFileStarted{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 4:
			return EventKindFileProgress{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 5:
			return EventKindFileDownloaded{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 6:
			return EventKindFileUploaded{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 7:
			return EventKindFileFailed{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeStatusINSTANCE.Read(reader),
			};
		case 8:
			return EventKindFileRejected{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 9:
			return EventKindFilePaused{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 10:
			return EventKindFileThrottled{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 11:
			return EventKindFilePending{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 12:
			return EventKindTransferFinalized{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 13:
			return EventKindTransferFailed{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeStatusINSTANCE.Read(reader),
			};
		case 14:
			return EventKindTransferDeferred{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterTypeStatusINSTANCE.Read(reader),
			};
		case 15:
			return EventKindFinalizeChecksumStarted{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 16:
			return EventKindFinalizeChecksumFinished{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 17:
			return EventKindFinalizeChecksumProgress{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 18:
			return EventKindVerifyChecksumStarted{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 19:
			return EventKindVerifyChecksumFinished{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 20:
			return EventKindVerifyChecksumProgress{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 21:
			return EventKindRuntimeError{
				FfiConverterTypeStatusCodeINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeEventKind.Read()", id));
	}
}

func (FfiConverterTypeEventKind) Write(writer io.Writer, value EventKind) {
	switch variant_value := value.(type) {
		case EventKindRequestReceived:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Peer)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterSequenceTypeReceivedFileINSTANCE.Write(writer, variant_value.Files)
		case EventKindRequestQueued:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Peer)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterSequenceTypeQueuedFileINSTANCE.Write(writer, variant_value.Files)
		case EventKindFileStarted:
			writeInt32(writer, 3)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Transferred)
		case EventKindFileProgress:
			writeInt32(writer, 4)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Transferred)
		case EventKindFileDownloaded:
			writeInt32(writer, 5)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FinalPath)
		case EventKindFileUploaded:
			writeInt32(writer, 6)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
		case EventKindFileFailed:
			writeInt32(writer, 7)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterTypeStatusINSTANCE.Write(writer, variant_value.Status)
		case EventKindFileRejected:
			writeInt32(writer, 8)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.ByPeer)
		case EventKindFilePaused:
			writeInt32(writer, 9)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
		case EventKindFileThrottled:
			writeInt32(writer, 10)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Transferred)
		case EventKindFilePending:
			writeInt32(writer, 11)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
		case EventKindTransferFinalized:
			writeInt32(writer, 12)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.ByPeer)
		case EventKindTransferFailed:
			writeInt32(writer, 13)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterTypeStatusINSTANCE.Write(writer, variant_value.Status)
		case EventKindTransferDeferred:
			writeInt32(writer, 14)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Peer)
			FfiConverterTypeStatusINSTANCE.Write(writer, variant_value.Status)
		case EventKindFinalizeChecksumStarted:
			writeInt32(writer, 15)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Size)
		case EventKindFinalizeChecksumFinished:
			writeInt32(writer, 16)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
		case EventKindFinalizeChecksumProgress:
			writeInt32(writer, 17)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesChecksummed)
		case EventKindVerifyChecksumStarted:
			writeInt32(writer, 18)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.Size)
		case EventKindVerifyChecksumFinished:
			writeInt32(writer, 19)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
		case EventKindVerifyChecksumProgress:
			writeInt32(writer, 20)
			FfiConverterStringINSTANCE.Write(writer, variant_value.TransferId)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FileId)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesChecksummed)
		case EventKindRuntimeError:
			writeInt32(writer, 21)
			FfiConverterTypeStatusCodeINSTANCE.Write(writer, variant_value.Status)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeEventKind.Write", value))
	}
}

type FfiDestroyerTypeEventKind struct {}

func (_ FfiDestroyerTypeEventKind) Destroy(value EventKind) {
	value.Destroy()
}




// Description of incoming file states.
// Some states are considered **terminal**. Terminal states appear
// once and it is the final state. Other states might appear multiple times.
type IncomingPathStateKind interface {
	Destroy()
}
// The download was issued for this file and it will proceed when
// possible.
type IncomingPathStateKindPending struct {
	BaseDir string
}

func (e IncomingPathStateKindPending) Destroy() {
		FfiDestroyerString{}.Destroy(e.BaseDir);
}
// The file was started to be received. Contains the base
// directory of the file.
type IncomingPathStateKindStarted struct {
	BytesReceived uint64
}

func (e IncomingPathStateKindStarted) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.BytesReceived);
}
// Contains status code of failure. 
// This is a **terminal** state.
type IncomingPathStateKindFailed struct {
	Status StatusCode
	BytesReceived uint64
}

func (e IncomingPathStateKindFailed) Destroy() {
		FfiDestroyerTypeStatusCode{}.Destroy(e.Status);
		FfiDestroyerUint64{}.Destroy(e.BytesReceived);
}
// The file was successfully received and saved to the disk.
// Contains the final path of the file.
// This is a **terminal** state.
type IncomingPathStateKindCompleted struct {
	FinalPath string
}

func (e IncomingPathStateKindCompleted) Destroy() {
		FfiDestroyerString{}.Destroy(e.FinalPath);
}
// The file was rejected by the receiver. Contains indicator of
// who rejected the file.
// This is a **terminal** state.
type IncomingPathStateKindRejected struct {
	ByPeer bool
	BytesReceived uint64
}

func (e IncomingPathStateKindRejected) Destroy() {
		FfiDestroyerBool{}.Destroy(e.ByPeer);
		FfiDestroyerUint64{}.Destroy(e.BytesReceived);
}
// The file was paused due to recoverable errors. Most probably
// due to network availability.
type IncomingPathStateKindPaused struct {
	BytesReceived uint64
}

func (e IncomingPathStateKindPaused) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.BytesReceived);
}

type FfiConverterTypeIncomingPathStateKind struct {}

var FfiConverterTypeIncomingPathStateKindINSTANCE = FfiConverterTypeIncomingPathStateKind{}

func (c FfiConverterTypeIncomingPathStateKind) Lift(rb RustBufferI) IncomingPathStateKind {
	return LiftFromRustBuffer[IncomingPathStateKind](c, rb)
}

func (c FfiConverterTypeIncomingPathStateKind) Lower(value IncomingPathStateKind) RustBuffer {
	return LowerIntoRustBuffer[IncomingPathStateKind](c, value)
}
func (FfiConverterTypeIncomingPathStateKind) Read(reader io.Reader) IncomingPathStateKind {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return IncomingPathStateKindPending{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return IncomingPathStateKindStarted{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 3:
			return IncomingPathStateKindFailed{
				FfiConverterTypeStatusCodeINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 4:
			return IncomingPathStateKindCompleted{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 5:
			return IncomingPathStateKindRejected{
				FfiConverterBoolINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 6:
			return IncomingPathStateKindPaused{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeIncomingPathStateKind.Read()", id));
	}
}

func (FfiConverterTypeIncomingPathStateKind) Write(writer io.Writer, value IncomingPathStateKind) {
	switch variant_value := value.(type) {
		case IncomingPathStateKindPending:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.BaseDir)
		case IncomingPathStateKindStarted:
			writeInt32(writer, 2)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesReceived)
		case IncomingPathStateKindFailed:
			writeInt32(writer, 3)
			FfiConverterTypeStatusCodeINSTANCE.Write(writer, variant_value.Status)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesReceived)
		case IncomingPathStateKindCompleted:
			writeInt32(writer, 4)
			FfiConverterStringINSTANCE.Write(writer, variant_value.FinalPath)
		case IncomingPathStateKindRejected:
			writeInt32(writer, 5)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.ByPeer)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesReceived)
		case IncomingPathStateKindPaused:
			writeInt32(writer, 6)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesReceived)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeIncomingPathStateKind.Write", value))
	}
}

type FfiDestroyerTypeIncomingPathStateKind struct {}

func (_ FfiDestroyerTypeIncomingPathStateKind) Destroy(value IncomingPathStateKind) {
	value.Destroy()
}


// The commmon error type thrown from functions
type LibdropError struct {
	err error
}

func (err LibdropError) Error() string {
	return fmt.Sprintf("LibdropError: %s", err.err.Error())
}

func (err LibdropError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrLibdropErrorUnknown = fmt.Errorf("LibdropErrorUnknown")
var ErrLibdropErrorInvalidString = fmt.Errorf("LibdropErrorInvalidString")
var ErrLibdropErrorBadInput = fmt.Errorf("LibdropErrorBadInput")
var ErrLibdropErrorTransferCreate = fmt.Errorf("LibdropErrorTransferCreate")
var ErrLibdropErrorNotStarted = fmt.Errorf("LibdropErrorNotStarted")
var ErrLibdropErrorAddrInUse = fmt.Errorf("LibdropErrorAddrInUse")
var ErrLibdropErrorInstanceStart = fmt.Errorf("LibdropErrorInstanceStart")
var ErrLibdropErrorInstanceStop = fmt.Errorf("LibdropErrorInstanceStop")
var ErrLibdropErrorInvalidPrivkey = fmt.Errorf("LibdropErrorInvalidPrivkey")
var ErrLibdropErrorDbError = fmt.Errorf("LibdropErrorDbError")

// Variant structs
// Operation resulted to unknown error.
type LibdropErrorUnknown struct {
	message string
}
// Operation resulted to unknown error.
func NewLibdropErrorUnknown(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorUnknown{
		},
	}
}

func (err LibdropErrorUnknown) Error() string {
	return fmt.Sprintf("Unknown: %s", err.message)
}

func (self LibdropErrorUnknown) Is(target error) bool {
	return target == ErrLibdropErrorUnknown
}
// The string provided is not valid UTF8
type LibdropErrorInvalidString struct {
	message string
}
// The string provided is not valid UTF8
func NewLibdropErrorInvalidString(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorInvalidString{
		},
	}
}

func (err LibdropErrorInvalidString) Error() string {
	return fmt.Sprintf("InvalidString: %s", err.message)
}

func (self LibdropErrorInvalidString) Is(target error) bool {
	return target == ErrLibdropErrorInvalidString
}
// One of the arguments provided is invalid
type LibdropErrorBadInput struct {
	message string
}
// One of the arguments provided is invalid
func NewLibdropErrorBadInput(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorBadInput{
		},
	}
}

func (err LibdropErrorBadInput) Error() string {
	return fmt.Sprintf("BadInput: %s", err.message)
}

func (self LibdropErrorBadInput) Is(target error) bool {
	return target == ErrLibdropErrorBadInput
}
// Failed to create transfer based on arguments provided
type LibdropErrorTransferCreate struct {
	message string
}
// Failed to create transfer based on arguments provided
func NewLibdropErrorTransferCreate(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorTransferCreate{
		},
	}
}

func (err LibdropErrorTransferCreate) Error() string {
	return fmt.Sprintf("TransferCreate: %s", err.message)
}

func (self LibdropErrorTransferCreate) Is(target error) bool {
	return target == ErrLibdropErrorTransferCreate
}
// The libdrop instance is not started yet
type LibdropErrorNotStarted struct {
	message string
}
// The libdrop instance is not started yet
func NewLibdropErrorNotStarted(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorNotStarted{
		},
	}
}

func (err LibdropErrorNotStarted) Error() string {
	return fmt.Sprintf("NotStarted: %s", err.message)
}

func (self LibdropErrorNotStarted) Is(target error) bool {
	return target == ErrLibdropErrorNotStarted
}
// Address already in use
type LibdropErrorAddrInUse struct {
	message string
}
// Address already in use
func NewLibdropErrorAddrInUse(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorAddrInUse{
		},
	}
}

func (err LibdropErrorAddrInUse) Error() string {
	return fmt.Sprintf("AddrInUse: %s", err.message)
}

func (self LibdropErrorAddrInUse) Is(target error) bool {
	return target == ErrLibdropErrorAddrInUse
}
// Failed to start the libdrop instance
type LibdropErrorInstanceStart struct {
	message string
}
// Failed to start the libdrop instance
func NewLibdropErrorInstanceStart(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorInstanceStart{
		},
	}
}

func (err LibdropErrorInstanceStart) Error() string {
	return fmt.Sprintf("InstanceStart: %s", err.message)
}

func (self LibdropErrorInstanceStart) Is(target error) bool {
	return target == ErrLibdropErrorInstanceStart
}
// Failed to stop the libdrop instance
type LibdropErrorInstanceStop struct {
	message string
}
// Failed to stop the libdrop instance
func NewLibdropErrorInstanceStop(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorInstanceStop{
		},
	}
}

func (err LibdropErrorInstanceStop) Error() string {
	return fmt.Sprintf("InstanceStop: %s", err.message)
}

func (self LibdropErrorInstanceStop) Is(target error) bool {
	return target == ErrLibdropErrorInstanceStop
}
// Invalid private key provided
type LibdropErrorInvalidPrivkey struct {
	message string
}
// Invalid private key provided
func NewLibdropErrorInvalidPrivkey(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorInvalidPrivkey{
		},
	}
}

func (err LibdropErrorInvalidPrivkey) Error() string {
	return fmt.Sprintf("InvalidPrivkey: %s", err.message)
}

func (self LibdropErrorInvalidPrivkey) Is(target error) bool {
	return target == ErrLibdropErrorInvalidPrivkey
}
// Database error
type LibdropErrorDbError struct {
	message string
}
// Database error
func NewLibdropErrorDbError(
) *LibdropError {
	return &LibdropError{
		err: &LibdropErrorDbError{
		},
	}
}

func (err LibdropErrorDbError) Error() string {
	return fmt.Sprintf("DbError: %s", err.message)
}

func (self LibdropErrorDbError) Is(target error) bool {
	return target == ErrLibdropErrorDbError
}

type FfiConverterTypeLibdropError struct{}

var FfiConverterTypeLibdropErrorINSTANCE = FfiConverterTypeLibdropError{}

func (c FfiConverterTypeLibdropError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeLibdropError) Lower(value *LibdropError) RustBuffer {
	return LowerIntoRustBuffer[*LibdropError](c, value)
}

func (c FfiConverterTypeLibdropError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &LibdropError{&LibdropErrorUnknown{message}}
	case 2:
		return &LibdropError{&LibdropErrorInvalidString{message}}
	case 3:
		return &LibdropError{&LibdropErrorBadInput{message}}
	case 4:
		return &LibdropError{&LibdropErrorTransferCreate{message}}
	case 5:
		return &LibdropError{&LibdropErrorNotStarted{message}}
	case 6:
		return &LibdropError{&LibdropErrorAddrInUse{message}}
	case 7:
		return &LibdropError{&LibdropErrorInstanceStart{message}}
	case 8:
		return &LibdropError{&LibdropErrorInstanceStop{message}}
	case 9:
		return &LibdropError{&LibdropErrorInvalidPrivkey{message}}
	case 10:
		return &LibdropError{&LibdropErrorDbError{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeLibdropError.Read()", errorID))
	}

	
}

func (c FfiConverterTypeLibdropError) Write(writer io.Writer, value *LibdropError) {
	switch variantValue := value.err.(type) {
		case *LibdropErrorUnknown:
			writeInt32(writer, 1)
		case *LibdropErrorInvalidString:
			writeInt32(writer, 2)
		case *LibdropErrorBadInput:
			writeInt32(writer, 3)
		case *LibdropErrorTransferCreate:
			writeInt32(writer, 4)
		case *LibdropErrorNotStarted:
			writeInt32(writer, 5)
		case *LibdropErrorAddrInUse:
			writeInt32(writer, 6)
		case *LibdropErrorInstanceStart:
			writeInt32(writer, 7)
		case *LibdropErrorInstanceStop:
			writeInt32(writer, 8)
		case *LibdropErrorInvalidPrivkey:
			writeInt32(writer, 9)
		case *LibdropErrorDbError:
			writeInt32(writer, 10)
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeLibdropError.Write", value))
	}
}



// Posible log levels.
type LogLevel uint

const (
	LogLevelCritical LogLevel = 1
	LogLevelError LogLevel = 2
	LogLevelWarning LogLevel = 3
	LogLevelInfo LogLevel = 4
	LogLevelDebug LogLevel = 5
	LogLevelTrace LogLevel = 6
)

type FfiConverterTypeLogLevel struct {}

var FfiConverterTypeLogLevelINSTANCE = FfiConverterTypeLogLevel{}

func (c FfiConverterTypeLogLevel) Lift(rb RustBufferI) LogLevel {
	return LiftFromRustBuffer[LogLevel](c, rb)
}

func (c FfiConverterTypeLogLevel) Lower(value LogLevel) RustBuffer {
	return LowerIntoRustBuffer[LogLevel](c, value)
}
func (FfiConverterTypeLogLevel) Read(reader io.Reader) LogLevel {
	id := readInt32(reader)
	return LogLevel(id)
}

func (FfiConverterTypeLogLevel) Write(writer io.Writer, value LogLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeLogLevel struct {}

func (_ FfiDestroyerTypeLogLevel) Destroy(value LogLevel) {
}




// The outgoing file data source
type OutgoingFileSource interface {
	Destroy()
}
// The file is read from disk, from the given path
type OutgoingFileSourceBasePath struct {
	BasePath string
}

func (e OutgoingFileSourceBasePath) Destroy() {
		FfiDestroyerString{}.Destroy(e.BasePath);
}
// The file descriptor is retrieved with the FD resolver
//
// # Warning
// This mechanism can only be used on UNIX systems
type OutgoingFileSourceContentUri struct {
	Uri string
}

func (e OutgoingFileSourceContentUri) Destroy() {
		FfiDestroyerString{}.Destroy(e.Uri);
}

type FfiConverterTypeOutgoingFileSource struct {}

var FfiConverterTypeOutgoingFileSourceINSTANCE = FfiConverterTypeOutgoingFileSource{}

func (c FfiConverterTypeOutgoingFileSource) Lift(rb RustBufferI) OutgoingFileSource {
	return LiftFromRustBuffer[OutgoingFileSource](c, rb)
}

func (c FfiConverterTypeOutgoingFileSource) Lower(value OutgoingFileSource) RustBuffer {
	return LowerIntoRustBuffer[OutgoingFileSource](c, value)
}
func (FfiConverterTypeOutgoingFileSource) Read(reader io.Reader) OutgoingFileSource {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return OutgoingFileSourceBasePath{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return OutgoingFileSourceContentUri{
				FfiConverterStringINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeOutgoingFileSource.Read()", id));
	}
}

func (FfiConverterTypeOutgoingFileSource) Write(writer io.Writer, value OutgoingFileSource) {
	switch variant_value := value.(type) {
		case OutgoingFileSourceBasePath:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.BasePath)
		case OutgoingFileSourceContentUri:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Uri)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeOutgoingFileSource.Write", value))
	}
}

type FfiDestroyerTypeOutgoingFileSource struct {}

func (_ FfiDestroyerTypeOutgoingFileSource) Destroy(value OutgoingFileSource) {
	value.Destroy()
}




// Description of outgoing file states.
// Some states are considered **terminal**. Terminal states appear
// once and it is the final state. Other states might appear multiple times.
type OutgoingPathStateKind interface {
	Destroy()
}
// The file was started to be received. Contains the base
// directory of the file.
type OutgoingPathStateKindStarted struct {
	BytesSent uint64
}

func (e OutgoingPathStateKindStarted) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.BytesSent);
}
// Contains status code of failure.
// This is a **terminal** state.
type OutgoingPathStateKindFailed struct {
	Status StatusCode
	BytesSent uint64
}

func (e OutgoingPathStateKindFailed) Destroy() {
		FfiDestroyerTypeStatusCode{}.Destroy(e.Status);
		FfiDestroyerUint64{}.Destroy(e.BytesSent);
}
// The file was successfully received and saved to the disk.
// Contains the final path of the file.
// This is a **terminal** state.
type OutgoingPathStateKindCompleted struct {
}

func (e OutgoingPathStateKindCompleted) Destroy() {
}
// The file was rejected by the receiver. Contains indicator of
// who rejected the file.
// This is a **terminal** state.
type OutgoingPathStateKindRejected struct {
	ByPeer bool
	BytesSent uint64
}

func (e OutgoingPathStateKindRejected) Destroy() {
		FfiDestroyerBool{}.Destroy(e.ByPeer);
		FfiDestroyerUint64{}.Destroy(e.BytesSent);
}
// The file was paused due to recoverable errors. Most probably
// due to network availability.
type OutgoingPathStateKindPaused struct {
	BytesSent uint64
}

func (e OutgoingPathStateKindPaused) Destroy() {
		FfiDestroyerUint64{}.Destroy(e.BytesSent);
}

type FfiConverterTypeOutgoingPathStateKind struct {}

var FfiConverterTypeOutgoingPathStateKindINSTANCE = FfiConverterTypeOutgoingPathStateKind{}

func (c FfiConverterTypeOutgoingPathStateKind) Lift(rb RustBufferI) OutgoingPathStateKind {
	return LiftFromRustBuffer[OutgoingPathStateKind](c, rb)
}

func (c FfiConverterTypeOutgoingPathStateKind) Lower(value OutgoingPathStateKind) RustBuffer {
	return LowerIntoRustBuffer[OutgoingPathStateKind](c, value)
}
func (FfiConverterTypeOutgoingPathStateKind) Read(reader io.Reader) OutgoingPathStateKind {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return OutgoingPathStateKindStarted{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 2:
			return OutgoingPathStateKindFailed{
				FfiConverterTypeStatusCodeINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 3:
			return OutgoingPathStateKindCompleted{
			};
		case 4:
			return OutgoingPathStateKindRejected{
				FfiConverterBoolINSTANCE.Read(reader),
				FfiConverterUint64INSTANCE.Read(reader),
			};
		case 5:
			return OutgoingPathStateKindPaused{
				FfiConverterUint64INSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeOutgoingPathStateKind.Read()", id));
	}
}

func (FfiConverterTypeOutgoingPathStateKind) Write(writer io.Writer, value OutgoingPathStateKind) {
	switch variant_value := value.(type) {
		case OutgoingPathStateKindStarted:
			writeInt32(writer, 1)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesSent)
		case OutgoingPathStateKindFailed:
			writeInt32(writer, 2)
			FfiConverterTypeStatusCodeINSTANCE.Write(writer, variant_value.Status)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesSent)
		case OutgoingPathStateKindCompleted:
			writeInt32(writer, 3)
		case OutgoingPathStateKindRejected:
			writeInt32(writer, 4)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.ByPeer)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesSent)
		case OutgoingPathStateKindPaused:
			writeInt32(writer, 5)
			FfiConverterUint64INSTANCE.Write(writer, variant_value.BytesSent)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeOutgoingPathStateKind.Write", value))
	}
}

type FfiDestroyerTypeOutgoingPathStateKind struct {}

func (_ FfiDestroyerTypeOutgoingPathStateKind) Destroy(value OutgoingPathStateKind) {
	value.Destroy()
}




// Status codes returend by the events
type StatusCode uint

const (
	// Not an error per se; indicates finalized transfers.
	StatusCodeFinalized StatusCode = 1
	// An invalid path was provided.
	// File path contains invalid components (e.g. parent `..`).
	StatusCodeBadPath StatusCode = 2
	// Failed to open the file or file doesn’t exist when asked to download. Might
	// indicate bad API usage. For Unix platforms using file descriptors, it might
	// indicate invalid FD being passed to libdrop.
	StatusCodeBadFile StatusCode = 3
	// Invalid input transfer ID passed.
	StatusCodeBadTransfer StatusCode = 4
	// An error occurred during the transfer and it cannot continue. The most probable
	// reason is the error occurred on the peer’s device or other error that cannot be
	// categorize elsewhere.
	StatusCodeBadTransferState StatusCode = 5
	// Invalid input file ID passed when.
	StatusCodeBadFileId StatusCode = 6
	// General IO error. Check the logs and contact libdrop team.
	StatusCodeIoError StatusCode = 7
	// Transfer limits exceeded. Limit is in terms of depth and breadth for
	// directories.
	StatusCodeTransferLimitsExceeded StatusCode = 8
	// The file size has changed since adding it to the transfer. The original file was
	// modified while not in flight in such a way that its size changed.
	StatusCodeMismatchedSize StatusCode = 9
	// An invalid argument was provided either as a function argument or
	// invalid config value.
	StatusCodeInvalidArgument StatusCode = 10
	// The WebSocket server failed to bind because of an address collision.
	StatusCodeAddrInUse StatusCode = 11
	// The file was modified while being uploaded.
	StatusCodeFileModified StatusCode = 12
	// The filename is too long which might be due to the fact the sender uses
	// a filesystem supporting longer filenames than the one which’s downloading the
	// file.
	StatusCodeFilenameTooLong StatusCode = 13
	// A peer couldn’t validate our authentication request.
	StatusCodeAuthenticationFailed StatusCode = 14
	// Persistence error.
	StatusCodeStorageError StatusCode = 15
	// The persistence database is lost. A new database will be created.
	StatusCodeDbLost StatusCode = 16
	// Downloaded file checksum differs from the advertised one. The downloaded
	// file is deleted by libdrop.
	StatusCodeFileChecksumMismatch StatusCode = 17
	// Download is impossible of the rejected file.
	StatusCodeFileRejected StatusCode = 18
	// Action is blocked because the failed condition has been reached.
	StatusCodeFileFailed StatusCode = 19
	// Action is blocked because the file is already transferred.
	StatusCodeFileFinished StatusCode = 20
	// Transfer requested with empty file list.
	StatusCodeEmptyTransfer StatusCode = 21
	// Transfer resume attempt was closed by peer for no reason. It might indicate
	// temporary issues on the peer’s side. It is safe to continue to resume the
	// transfer.
	StatusCodeConnectionClosedByPeer StatusCode = 22
	// Peer’s DDoS protection kicked in.
	// Transfer should be resumed after some cooldown period.
	StatusCodeTooManyRequests StatusCode = 23
	// This error code is intercepted from the OS errors. Indicate lack of
	// privileges to do certain operation.
	StatusCodePermissionDenied StatusCode = 24
)

type FfiConverterTypeStatusCode struct {}

var FfiConverterTypeStatusCodeINSTANCE = FfiConverterTypeStatusCode{}

func (c FfiConverterTypeStatusCode) Lift(rb RustBufferI) StatusCode {
	return LiftFromRustBuffer[StatusCode](c, rb)
}

func (c FfiConverterTypeStatusCode) Lower(value StatusCode) RustBuffer {
	return LowerIntoRustBuffer[StatusCode](c, value)
}
func (FfiConverterTypeStatusCode) Read(reader io.Reader) StatusCode {
	id := readInt32(reader)
	return StatusCode(id)
}

func (FfiConverterTypeStatusCode) Write(writer io.Writer, value StatusCode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeStatusCode struct {}

func (_ FfiDestroyerTypeStatusCode) Destroy(value StatusCode) {
}




// The transfer file description
type TransferDescriptor interface {
	Destroy()
}
// Disk file with the given path
type TransferDescriptorPath struct {
	Path string
}

func (e TransferDescriptorPath) Destroy() {
		FfiDestroyerString{}.Destroy(e.Path);
}
// File descriptor with the given URI (used for the `FdResolver`)
type TransferDescriptorFd struct {
	Filename string
	ContentUri string
	Fd *int32
}

func (e TransferDescriptorFd) Destroy() {
		FfiDestroyerString{}.Destroy(e.Filename);
		FfiDestroyerString{}.Destroy(e.ContentUri);
		FfiDestroyerOptionalInt32{}.Destroy(e.Fd);
}

type FfiConverterTypeTransferDescriptor struct {}

var FfiConverterTypeTransferDescriptorINSTANCE = FfiConverterTypeTransferDescriptor{}

func (c FfiConverterTypeTransferDescriptor) Lift(rb RustBufferI) TransferDescriptor {
	return LiftFromRustBuffer[TransferDescriptor](c, rb)
}

func (c FfiConverterTypeTransferDescriptor) Lower(value TransferDescriptor) RustBuffer {
	return LowerIntoRustBuffer[TransferDescriptor](c, value)
}
func (FfiConverterTypeTransferDescriptor) Read(reader io.Reader) TransferDescriptor {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TransferDescriptorPath{
				FfiConverterStringINSTANCE.Read(reader),
			};
		case 2:
			return TransferDescriptorFd{
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterStringINSTANCE.Read(reader),
				FfiConverterOptionalInt32INSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTransferDescriptor.Read()", id));
	}
}

func (FfiConverterTypeTransferDescriptor) Write(writer io.Writer, value TransferDescriptor) {
	switch variant_value := value.(type) {
		case TransferDescriptorPath:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Path)
		case TransferDescriptorFd:
			writeInt32(writer, 2)
			FfiConverterStringINSTANCE.Write(writer, variant_value.Filename)
			FfiConverterStringINSTANCE.Write(writer, variant_value.ContentUri)
			FfiConverterOptionalInt32INSTANCE.Write(writer, variant_value.Fd)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTransferDescriptor.Write", value))
	}
}

type FfiDestroyerTypeTransferDescriptor struct {}

func (_ FfiDestroyerTypeTransferDescriptor) Destroy(value TransferDescriptor) {
	value.Destroy()
}




// A type of the transfer
type TransferKind interface {
	Destroy()
}
// The transfer is incoming, meaning we are the one who receives the files
type TransferKindIncoming struct {
	Paths []IncomingPath
}

func (e TransferKindIncoming) Destroy() {
		FfiDestroyerSequenceTypeIncomingPath{}.Destroy(e.Paths);
}
// The transfer is incoming, meaning we are the one who sends the files
type TransferKindOutgoing struct {
	Paths []OutgoingPath
}

func (e TransferKindOutgoing) Destroy() {
		FfiDestroyerSequenceTypeOutgoingPath{}.Destroy(e.Paths);
}

type FfiConverterTypeTransferKind struct {}

var FfiConverterTypeTransferKindINSTANCE = FfiConverterTypeTransferKind{}

func (c FfiConverterTypeTransferKind) Lift(rb RustBufferI) TransferKind {
	return LiftFromRustBuffer[TransferKind](c, rb)
}

func (c FfiConverterTypeTransferKind) Lower(value TransferKind) RustBuffer {
	return LowerIntoRustBuffer[TransferKind](c, value)
}
func (FfiConverterTypeTransferKind) Read(reader io.Reader) TransferKind {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TransferKindIncoming{
				FfiConverterSequenceTypeIncomingPathINSTANCE.Read(reader),
			};
		case 2:
			return TransferKindOutgoing{
				FfiConverterSequenceTypeOutgoingPathINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTransferKind.Read()", id));
	}
}

func (FfiConverterTypeTransferKind) Write(writer io.Writer, value TransferKind) {
	switch variant_value := value.(type) {
		case TransferKindIncoming:
			writeInt32(writer, 1)
			FfiConverterSequenceTypeIncomingPathINSTANCE.Write(writer, variant_value.Paths)
		case TransferKindOutgoing:
			writeInt32(writer, 2)
			FfiConverterSequenceTypeOutgoingPathINSTANCE.Write(writer, variant_value.Paths)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTransferKind.Write", value))
	}
}

type FfiDestroyerTypeTransferKind struct {}

func (_ FfiDestroyerTypeTransferKind) Destroy(value TransferKind) {
	value.Destroy()
}




// Description of the transfer state
type TransferStateKind interface {
	Destroy()
}
// The transfer was successfully canceled by either peer.
// Contains indicator of who canceled the transfer.
type TransferStateKindCancel struct {
	ByPeer bool
}

func (e TransferStateKindCancel) Destroy() {
		FfiDestroyerBool{}.Destroy(e.ByPeer);
}
// Contains status code of failure.
type TransferStateKindFailed struct {
	Status StatusCode
}

func (e TransferStateKindFailed) Destroy() {
		FfiDestroyerTypeStatusCode{}.Destroy(e.Status);
}

type FfiConverterTypeTransferStateKind struct {}

var FfiConverterTypeTransferStateKindINSTANCE = FfiConverterTypeTransferStateKind{}

func (c FfiConverterTypeTransferStateKind) Lift(rb RustBufferI) TransferStateKind {
	return LiftFromRustBuffer[TransferStateKind](c, rb)
}

func (c FfiConverterTypeTransferStateKind) Lower(value TransferStateKind) RustBuffer {
	return LowerIntoRustBuffer[TransferStateKind](c, value)
}
func (FfiConverterTypeTransferStateKind) Read(reader io.Reader) TransferStateKind {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return TransferStateKindCancel{
				FfiConverterBoolINSTANCE.Read(reader),
			};
		case 2:
			return TransferStateKindFailed{
				FfiConverterTypeStatusCodeINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeTransferStateKind.Read()", id));
	}
}

func (FfiConverterTypeTransferStateKind) Write(writer io.Writer, value TransferStateKind) {
	switch variant_value := value.(type) {
		case TransferStateKindCancel:
			writeInt32(writer, 1)
			FfiConverterBoolINSTANCE.Write(writer, variant_value.ByPeer)
		case TransferStateKindFailed:
			writeInt32(writer, 2)
			FfiConverterTypeStatusCodeINSTANCE.Write(writer, variant_value.Status)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeTransferStateKind.Write", value))
	}
}

type FfiDestroyerTypeTransferStateKind struct {}

func (_ FfiDestroyerTypeTransferStateKind) Destroy(value TransferStateKind) {
	value.Destroy()
}




type uniffiCallbackResult C.int32_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)


type concurrentHandleMap[T any] struct {
	leftMap       map[uint64]*T
	rightMap      map[*T]uint64
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		leftMap:  map[uint64]*T{},
		rightMap: map[*T]uint64{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj *T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if existingHandle, ok := cm.rightMap[obj]; ok {
		return existingHandle
	}
	cm.currentHandle = cm.currentHandle + 1
	cm.leftMap[cm.currentHandle] = obj
	cm.rightMap[obj] = cm.currentHandle
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if val, ok := cm.leftMap[handle]; ok {
		delete(cm.leftMap, handle)
		delete(cm.rightMap, val)
	}
	return false
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (*T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.leftMap[handle]
	return val, ok
}

type FfiConverterCallbackInterface[CallbackInterface any] struct {
	handleMap *concurrentHandleMap[CallbackInterface]
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lift(handle uint64) CallbackInterface {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return *val
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Read(reader io.Reader) CallbackInterface {
	return c.Lift(readUint64(reader))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lower(value CallbackInterface) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(&value))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Write(writer io.Writer, value CallbackInterface) {
	writeUint64(writer, uint64(c.Lower(value)))
}
// The event callback
type EventCallback interface {
	
	// Method called whenever event occurs
	OnEvent(event Event) 
	
}

// foreignCallbackCallbackInterfaceEventCallback cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceEventCallback struct {}

//export norddrop_cgo_EventCallback
func norddrop_cgo_EventCallback(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceEventCallbackINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceEventCallbackINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceEventCallback{}.InvokeOnEvent(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceEventCallback) InvokeOnEvent (callback EventCallback, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.OnEvent(FfiConverterTypeEventINSTANCE.Read(reader));

        
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceEventCallback struct {
	FfiConverterCallbackInterface[EventCallback]
}

var FfiConverterCallbackInterfaceEventCallbackINSTANCE = &FfiConverterCallbackInterfaceEventCallback {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[EventCallback]{
		handleMap: newConcurrentHandleMap[EventCallback](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceEventCallback) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_norddrop_fn_init_callback_eventcallback(C.ForeignCallback(C.norddrop_cgo_EventCallback), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceEventCallback struct {}

func (FfiDestroyerCallbackInterfaceEventCallback) Destroy(value EventCallback) {
}





// Profides the file descriptor based on the content URI
//
// # Warning
// Can be used only on UNIX systems
type FdResolver interface {
	
	OnFd(contentUri string) *int32
	
}

// foreignCallbackCallbackInterfaceFdResolver cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceFdResolver struct {}

//export norddrop_cgo_FdResolver
func norddrop_cgo_FdResolver(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceFdResolverINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceFdResolverINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceFdResolver{}.InvokeOnFd(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceFdResolver) InvokeOnFd (callback FdResolver, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result :=callback.OnFd(FfiConverterStringINSTANCE.Read(reader));

        
	*outBuf = LowerIntoRustBuffer[*int32](FfiConverterOptionalInt32INSTANCE, result)
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceFdResolver struct {
	FfiConverterCallbackInterface[FdResolver]
}

var FfiConverterCallbackInterfaceFdResolverINSTANCE = &FfiConverterCallbackInterfaceFdResolver {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[FdResolver]{
		handleMap: newConcurrentHandleMap[FdResolver](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceFdResolver) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_norddrop_fn_init_callback_fdresolver(C.ForeignCallback(C.norddrop_cgo_FdResolver), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceFdResolver struct {}

func (FfiDestroyerCallbackInterfaceFdResolver) Destroy(value FdResolver) {
}





// The interface for providing crypto keys
type KeyStore interface {
	
	// It is used to request
	// the app to provide the peer’s public key or the node itself. 
	//
	// # Arguments
	// * `peer` - peer's IP address
	//
	// # Returns
	// 32bytes private key. Note that it’s not BASE64, it must
	// be decoded if it is beforehand.
	// The `null` value is used to indicate that the key could not be
	// provided.
	OnPubkey(peer string) *[]byte
	
	// 32bytes private key
	//
	// # Warning
	// This it’s not BASE64, it must
	// be decoded if it is beforehand.
	Privkey() []byte
	
}

// foreignCallbackCallbackInterfaceKeyStore cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceKeyStore struct {}

//export norddrop_cgo_KeyStore
func norddrop_cgo_KeyStore(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceKeyStoreINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceKeyStoreINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceKeyStore{}.InvokeOnPubkey(cb, args, outBuf);
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceKeyStore{}.InvokePrivkey(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceKeyStore) InvokeOnPubkey (callback KeyStore, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result :=callback.OnPubkey(FfiConverterStringINSTANCE.Read(reader));

        
	*outBuf = LowerIntoRustBuffer[*[]byte](FfiConverterOptionalBytesINSTANCE, result)
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceKeyStore) InvokePrivkey (callback KeyStore, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	result :=callback.Privkey();

        
	*outBuf = LowerIntoRustBuffer[[]byte](FfiConverterBytesINSTANCE, result)
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceKeyStore struct {
	FfiConverterCallbackInterface[KeyStore]
}

var FfiConverterCallbackInterfaceKeyStoreINSTANCE = &FfiConverterCallbackInterfaceKeyStore {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[KeyStore]{
		handleMap: newConcurrentHandleMap[KeyStore](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceKeyStore) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_norddrop_fn_init_callback_keystore(C.ForeignCallback(C.norddrop_cgo_KeyStore), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceKeyStore struct {}

func (FfiDestroyerCallbackInterfaceKeyStore) Destroy(value KeyStore) {
}





// The logger callback interface
type Logger interface {
	
	// Function called when log message occurs
	OnLog(level LogLevel, msg string) 
	
	// Maximum log level
	Level() LogLevel
	
}

// foreignCallbackCallbackInterfaceLogger cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceLogger struct {}

//export norddrop_cgo_Logger
func norddrop_cgo_Logger(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceLoggerINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceLoggerINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceLogger{}.InvokeOnLog(cb, args, outBuf);
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceLogger{}.InvokeLevel(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceLogger) InvokeOnLog (callback Logger, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.OnLog(FfiConverterTypeLogLevelINSTANCE.Read(reader), FfiConverterStringINSTANCE.Read(reader));

        
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceLogger) InvokeLevel (callback Logger, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	result :=callback.Level();

        
	*outBuf = LowerIntoRustBuffer[LogLevel](FfiConverterTypeLogLevelINSTANCE, result)
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceLogger struct {
	FfiConverterCallbackInterface[Logger]
}

var FfiConverterCallbackInterfaceLoggerINSTANCE = &FfiConverterCallbackInterfaceLogger {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[Logger]{
		handleMap: newConcurrentHandleMap[Logger](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceLogger) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_norddrop_fn_init_callback_logger(C.ForeignCallback(C.norddrop_cgo_Logger), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceLogger struct {}

func (FfiDestroyerCallbackInterfaceLogger) Destroy(value Logger) {
}




type FfiConverterOptionalUint32 struct{}

var FfiConverterOptionalUint32INSTANCE = FfiConverterOptionalUint32{}

func (c FfiConverterOptionalUint32) Lift(rb RustBufferI) *uint32 {
	return LiftFromRustBuffer[*uint32](c, rb)
}

func (_ FfiConverterOptionalUint32) Read(reader io.Reader) *uint32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint32) Lower(value *uint32) RustBuffer {
	return LowerIntoRustBuffer[*uint32](c, value)
}

func (_ FfiConverterOptionalUint32) Write(writer io.Writer, value *uint32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint32 struct {}

func (_ FfiDestroyerOptionalUint32) Destroy(value *uint32) {
	if value != nil {
		FfiDestroyerUint32{}.Destroy(*value)
	}
}



type FfiConverterOptionalInt32 struct{}

var FfiConverterOptionalInt32INSTANCE = FfiConverterOptionalInt32{}

func (c FfiConverterOptionalInt32) Lift(rb RustBufferI) *int32 {
	return LiftFromRustBuffer[*int32](c, rb)
}

func (_ FfiConverterOptionalInt32) Read(reader io.Reader) *int32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterInt32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalInt32) Lower(value *int32) RustBuffer {
	return LowerIntoRustBuffer[*int32](c, value)
}

func (_ FfiConverterOptionalInt32) Write(writer io.Writer, value *int32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterInt32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalInt32 struct {}

func (_ FfiDestroyerOptionalInt32) Destroy(value *int32) {
	if value != nil {
		FfiDestroyerInt32{}.Destroy(*value)
	}
}



type FfiConverterOptionalUint64 struct{}

var FfiConverterOptionalUint64INSTANCE = FfiConverterOptionalUint64{}

func (c FfiConverterOptionalUint64) Lift(rb RustBufferI) *uint64 {
	return LiftFromRustBuffer[*uint64](c, rb)
}

func (_ FfiConverterOptionalUint64) Read(reader io.Reader) *uint64 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint64INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint64) Lower(value *uint64) RustBuffer {
	return LowerIntoRustBuffer[*uint64](c, value)
}

func (_ FfiConverterOptionalUint64) Write(writer io.Writer, value *uint64) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint64 struct {}

func (_ FfiDestroyerOptionalUint64) Destroy(value *uint64) {
	if value != nil {
		FfiDestroyerUint64{}.Destroy(*value)
	}
}



type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct {}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}



type FfiConverterOptionalBytes struct{}

var FfiConverterOptionalBytesINSTANCE = FfiConverterOptionalBytes{}

func (c FfiConverterOptionalBytes) Lift(rb RustBufferI) *[]byte {
	return LiftFromRustBuffer[*[]byte](c, rb)
}

func (_ FfiConverterOptionalBytes) Read(reader io.Reader) *[]byte {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterBytesINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalBytes) Lower(value *[]byte) RustBuffer {
	return LowerIntoRustBuffer[*[]byte](c, value)
}

func (_ FfiConverterOptionalBytes) Write(writer io.Writer, value *[]byte) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterBytesINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalBytes struct {}

func (_ FfiDestroyerOptionalBytes) Destroy(value *[]byte) {
	if value != nil {
		FfiDestroyerBytes{}.Destroy(*value)
	}
}



type FfiConverterSequenceString struct{}

var FfiConverterSequenceStringINSTANCE = FfiConverterSequenceString{}

func (c FfiConverterSequenceString) Lift(rb RustBufferI) []string {
	return LiftFromRustBuffer[[]string](c, rb)
}

func (c FfiConverterSequenceString) Read(reader io.Reader) []string {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]string, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterStringINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceString) Lower(value []string) RustBuffer {
	return LowerIntoRustBuffer[[]string](c, value)
}

func (c FfiConverterSequenceString) Write(writer io.Writer, value []string) {
	if len(value) > math.MaxInt32 {
		panic("[]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterStringINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceString struct {}

func (FfiDestroyerSequenceString) Destroy(sequence []string) {
	for _, value := range sequence {
		FfiDestroyerString{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeIncomingPath struct{}

var FfiConverterSequenceTypeIncomingPathINSTANCE = FfiConverterSequenceTypeIncomingPath{}

func (c FfiConverterSequenceTypeIncomingPath) Lift(rb RustBufferI) []IncomingPath {
	return LiftFromRustBuffer[[]IncomingPath](c, rb)
}

func (c FfiConverterSequenceTypeIncomingPath) Read(reader io.Reader) []IncomingPath {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]IncomingPath, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeIncomingPathINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeIncomingPath) Lower(value []IncomingPath) RustBuffer {
	return LowerIntoRustBuffer[[]IncomingPath](c, value)
}

func (c FfiConverterSequenceTypeIncomingPath) Write(writer io.Writer, value []IncomingPath) {
	if len(value) > math.MaxInt32 {
		panic("[]IncomingPath is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeIncomingPathINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeIncomingPath struct {}

func (FfiDestroyerSequenceTypeIncomingPath) Destroy(sequence []IncomingPath) {
	for _, value := range sequence {
		FfiDestroyerTypeIncomingPath{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeIncomingPathState struct{}

var FfiConverterSequenceTypeIncomingPathStateINSTANCE = FfiConverterSequenceTypeIncomingPathState{}

func (c FfiConverterSequenceTypeIncomingPathState) Lift(rb RustBufferI) []IncomingPathState {
	return LiftFromRustBuffer[[]IncomingPathState](c, rb)
}

func (c FfiConverterSequenceTypeIncomingPathState) Read(reader io.Reader) []IncomingPathState {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]IncomingPathState, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeIncomingPathStateINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeIncomingPathState) Lower(value []IncomingPathState) RustBuffer {
	return LowerIntoRustBuffer[[]IncomingPathState](c, value)
}

func (c FfiConverterSequenceTypeIncomingPathState) Write(writer io.Writer, value []IncomingPathState) {
	if len(value) > math.MaxInt32 {
		panic("[]IncomingPathState is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeIncomingPathStateINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeIncomingPathState struct {}

func (FfiDestroyerSequenceTypeIncomingPathState) Destroy(sequence []IncomingPathState) {
	for _, value := range sequence {
		FfiDestroyerTypeIncomingPathState{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeOutgoingPath struct{}

var FfiConverterSequenceTypeOutgoingPathINSTANCE = FfiConverterSequenceTypeOutgoingPath{}

func (c FfiConverterSequenceTypeOutgoingPath) Lift(rb RustBufferI) []OutgoingPath {
	return LiftFromRustBuffer[[]OutgoingPath](c, rb)
}

func (c FfiConverterSequenceTypeOutgoingPath) Read(reader io.Reader) []OutgoingPath {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]OutgoingPath, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeOutgoingPathINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeOutgoingPath) Lower(value []OutgoingPath) RustBuffer {
	return LowerIntoRustBuffer[[]OutgoingPath](c, value)
}

func (c FfiConverterSequenceTypeOutgoingPath) Write(writer io.Writer, value []OutgoingPath) {
	if len(value) > math.MaxInt32 {
		panic("[]OutgoingPath is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeOutgoingPathINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeOutgoingPath struct {}

func (FfiDestroyerSequenceTypeOutgoingPath) Destroy(sequence []OutgoingPath) {
	for _, value := range sequence {
		FfiDestroyerTypeOutgoingPath{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeOutgoingPathState struct{}

var FfiConverterSequenceTypeOutgoingPathStateINSTANCE = FfiConverterSequenceTypeOutgoingPathState{}

func (c FfiConverterSequenceTypeOutgoingPathState) Lift(rb RustBufferI) []OutgoingPathState {
	return LiftFromRustBuffer[[]OutgoingPathState](c, rb)
}

func (c FfiConverterSequenceTypeOutgoingPathState) Read(reader io.Reader) []OutgoingPathState {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]OutgoingPathState, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeOutgoingPathStateINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeOutgoingPathState) Lower(value []OutgoingPathState) RustBuffer {
	return LowerIntoRustBuffer[[]OutgoingPathState](c, value)
}

func (c FfiConverterSequenceTypeOutgoingPathState) Write(writer io.Writer, value []OutgoingPathState) {
	if len(value) > math.MaxInt32 {
		panic("[]OutgoingPathState is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeOutgoingPathStateINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeOutgoingPathState struct {}

func (FfiDestroyerSequenceTypeOutgoingPathState) Destroy(sequence []OutgoingPathState) {
	for _, value := range sequence {
		FfiDestroyerTypeOutgoingPathState{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeQueuedFile struct{}

var FfiConverterSequenceTypeQueuedFileINSTANCE = FfiConverterSequenceTypeQueuedFile{}

func (c FfiConverterSequenceTypeQueuedFile) Lift(rb RustBufferI) []QueuedFile {
	return LiftFromRustBuffer[[]QueuedFile](c, rb)
}

func (c FfiConverterSequenceTypeQueuedFile) Read(reader io.Reader) []QueuedFile {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]QueuedFile, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeQueuedFileINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeQueuedFile) Lower(value []QueuedFile) RustBuffer {
	return LowerIntoRustBuffer[[]QueuedFile](c, value)
}

func (c FfiConverterSequenceTypeQueuedFile) Write(writer io.Writer, value []QueuedFile) {
	if len(value) > math.MaxInt32 {
		panic("[]QueuedFile is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeQueuedFileINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeQueuedFile struct {}

func (FfiDestroyerSequenceTypeQueuedFile) Destroy(sequence []QueuedFile) {
	for _, value := range sequence {
		FfiDestroyerTypeQueuedFile{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeReceivedFile struct{}

var FfiConverterSequenceTypeReceivedFileINSTANCE = FfiConverterSequenceTypeReceivedFile{}

func (c FfiConverterSequenceTypeReceivedFile) Lift(rb RustBufferI) []ReceivedFile {
	return LiftFromRustBuffer[[]ReceivedFile](c, rb)
}

func (c FfiConverterSequenceTypeReceivedFile) Read(reader io.Reader) []ReceivedFile {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ReceivedFile, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeReceivedFileINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeReceivedFile) Lower(value []ReceivedFile) RustBuffer {
	return LowerIntoRustBuffer[[]ReceivedFile](c, value)
}

func (c FfiConverterSequenceTypeReceivedFile) Write(writer io.Writer, value []ReceivedFile) {
	if len(value) > math.MaxInt32 {
		panic("[]ReceivedFile is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeReceivedFileINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeReceivedFile struct {}

func (FfiDestroyerSequenceTypeReceivedFile) Destroy(sequence []ReceivedFile) {
	for _, value := range sequence {
		FfiDestroyerTypeReceivedFile{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTransferInfo struct{}

var FfiConverterSequenceTypeTransferInfoINSTANCE = FfiConverterSequenceTypeTransferInfo{}

func (c FfiConverterSequenceTypeTransferInfo) Lift(rb RustBufferI) []TransferInfo {
	return LiftFromRustBuffer[[]TransferInfo](c, rb)
}

func (c FfiConverterSequenceTypeTransferInfo) Read(reader io.Reader) []TransferInfo {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TransferInfo, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTransferInfoINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTransferInfo) Lower(value []TransferInfo) RustBuffer {
	return LowerIntoRustBuffer[[]TransferInfo](c, value)
}

func (c FfiConverterSequenceTypeTransferInfo) Write(writer io.Writer, value []TransferInfo) {
	if len(value) > math.MaxInt32 {
		panic("[]TransferInfo is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTransferInfoINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTransferInfo struct {}

func (FfiDestroyerSequenceTypeTransferInfo) Destroy(sequence []TransferInfo) {
	for _, value := range sequence {
		FfiDestroyerTypeTransferInfo{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTransferState struct{}

var FfiConverterSequenceTypeTransferStateINSTANCE = FfiConverterSequenceTypeTransferState{}

func (c FfiConverterSequenceTypeTransferState) Lift(rb RustBufferI) []TransferState {
	return LiftFromRustBuffer[[]TransferState](c, rb)
}

func (c FfiConverterSequenceTypeTransferState) Read(reader io.Reader) []TransferState {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TransferState, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTransferStateINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTransferState) Lower(value []TransferState) RustBuffer {
	return LowerIntoRustBuffer[[]TransferState](c, value)
}

func (c FfiConverterSequenceTypeTransferState) Write(writer io.Writer, value []TransferState) {
	if len(value) > math.MaxInt32 {
		panic("[]TransferState is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTransferStateINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTransferState struct {}

func (FfiDestroyerSequenceTypeTransferState) Destroy(sequence []TransferState) {
	for _, value := range sequence {
		FfiDestroyerTypeTransferState{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTransferDescriptor struct{}

var FfiConverterSequenceTypeTransferDescriptorINSTANCE = FfiConverterSequenceTypeTransferDescriptor{}

func (c FfiConverterSequenceTypeTransferDescriptor) Lift(rb RustBufferI) []TransferDescriptor {
	return LiftFromRustBuffer[[]TransferDescriptor](c, rb)
}

func (c FfiConverterSequenceTypeTransferDescriptor) Read(reader io.Reader) []TransferDescriptor {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TransferDescriptor, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTransferDescriptorINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTransferDescriptor) Lower(value []TransferDescriptor) RustBuffer {
	return LowerIntoRustBuffer[[]TransferDescriptor](c, value)
}

func (c FfiConverterSequenceTypeTransferDescriptor) Write(writer io.Writer, value []TransferDescriptor) {
	if len(value) > math.MaxInt32 {
		panic("[]TransferDescriptor is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTransferDescriptorINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTransferDescriptor struct {}

func (FfiDestroyerSequenceTypeTransferDescriptor) Destroy(sequence []TransferDescriptor) {
	for _, value := range sequence {
		FfiDestroyerTypeTransferDescriptor{}.Destroy(value)	
	}
}

// Returs the libdrop version
func Version() string {
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_norddrop_fn_func_version( _uniffiStatus)
	}))
}

