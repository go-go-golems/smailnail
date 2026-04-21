package annotationui

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "marshal proto json").Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// #nosec G705 -- body is JSON from protojson.Marshal, not user-controlled HTML/JS
	_, _ = w.Write(body)
}

func decodeProtoJSONBody(w http.ResponseWriter, r *http.Request, dest proto.Message) bool {
	defer func() {
		_ = r.Body.Close()
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, errors.Wrap(err, "read json body").Error())
		return false
	}
	if err := protoJSONUnmarshalOptions.Unmarshal(body, dest); err != nil {
		writeMessageError(w, http.StatusBadRequest, errors.Wrap(err, "decode proto json").Error())
		return false
	}
	return true
}
