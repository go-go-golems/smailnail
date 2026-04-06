package smailnaild

import (
	"io"
	"net/http"

	appv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/app/v1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var protoJSONMarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   false,
	UseEnumNumbers:  false,
	EmitUnpopulated: true,
}

var protoJSONUnmarshalOptions = protojson.UnmarshalOptions{
	DiscardUnknown: false,
}

func writeProtoJSON(w http.ResponseWriter, statusCode int, payload proto.Message) {
	body, err := protoJSONMarshalOptions.Marshal(payload)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal-error", errors.Wrap(err, "marshal proto json").Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func decodeProtoJSONBody(w http.ResponseWriter, r *http.Request, dest proto.Message) bool {
	defer func() {
		_ = r.Body.Close()
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", errors.Wrap(err, "read json body").Error(), nil)
		return false
	}
	if err := protoJSONUnmarshalOptions.Unmarshal(body, dest); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", errors.Wrap(err, "decode proto json").Error(), nil)
		return false
	}
	return true
}

func decodeProtoJSONBodyAllowEmpty(w http.ResponseWriter, r *http.Request, dest proto.Message) bool {
	if r.Body == nil {
		return true
	}
	defer func() {
		_ = r.Body.Close()
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", errors.Wrap(err, "read json body").Error(), nil)
		return false
	}
	if len(body) == 0 {
		return true
	}
	if err := protoJSONUnmarshalOptions.Unmarshal(body, dest); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", errors.Wrap(err, "decode proto json").Error(), nil)
		return false
	}
	return true
}

func writeProtoAPIError(w http.ResponseWriter, statusCode int, code, message string, details map[string]any) {
	payload := &appv1.ErrorResponse{
		Error: &appv1.ApiError{
			Code:    code,
			Message: message,
		},
	}
	if len(details) > 0 {
		structured, err := structpb.NewStruct(details)
		if err == nil {
			payload.Error.Details = structured
		}
	}
	writeProtoJSON(w, statusCode, payload)
}
