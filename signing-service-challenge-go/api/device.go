package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/dto"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
	"github.com/google/uuid"
)

type KeyPair interface {
	ToBytes() ([]byte, []byte, error)
}

func (s *Server) CreateSignatureDevice(response http.ResponseWriter, request *http.Request) {
	var payload dto.CreateSignatureDeviceRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"invalid payload"})

		return
	}

	id := uuid.NewString()
	privateKey, publicKey, err := getKeys(domain.SignatureAlgorithm(payload.Algorithm))
	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{fmt.Sprintf("failed to create keys: %s", err.Error())})

		return
	}

	err = s.store.Store(id, domain.SignatureDevice{
		ID:         id,
		Label:      payload.Label,
		Algorithm:  domain.SignatureAlgorithm(payload.Algorithm),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	})

	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{fmt.Sprintf("failed to store device: %s", err.Error())})

		return
	}

	WriteAPIResponse(response, http.StatusOK, dto.CreateSignatureDeviceResponse{ID: id})
}

func (s *Server) SignTransaction(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing device ID"})

		return
	}

	var payload dto.SignTransactionRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"invalid payload"})

		return
	}

	if payload.Data == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing data to be signed"})

		return
	}

	device, err := s.store.Get(id)
	if err != nil {
		handleStoreErr(response, err)

		return
	}

	if err = s.store.Lock(id); err != nil {
		handleStoreErr(response, err)

		return
	}

	defer s.store.Unlock(id)

	lastSignature := device.LastSignature
	if device.Counter == 0 {
		lastSignature = base64.StdEncoding.EncodeToString([]byte(device.ID))
	}

	dataToSign := fmt.Sprintf("%d_%s_%s", device.Counter, payload.Data, lastSignature)
	data, err := getSignedData(dataToSign, device.PrivateKey, device.Algorithm)
	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{
			fmt.Sprintf("error while signing data: %s", err.Error()),
		})

		return
	}

	signature := base64.StdEncoding.EncodeToString(data)

	device.Counter += 1
	device.LastSignature = signature

	err = s.store.Set(device.ID, *device)
	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{
			fmt.Sprintf("error while updating store data: %s", err.Error()),
		})

		return
	}

	WriteAPIResponse(response, http.StatusOK, dto.SignTransactionResponse{
		Signature:  signature,
		SignedData: dataToSign,
	})
}

func (s *Server) GetSignatureDevice(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing device ID"})

		return
	}

	device, err := s.store.Get(id)
	if err != nil {
		handleStoreErr(response, err)

		return
	}

	WriteAPIResponse(response, http.StatusOK, dto.GetDeviceResponse{
		ID:    device.ID,
		Label: device.Label,
	})
}

func (s *Server) ListSignatureDevice(response http.ResponseWriter, request *http.Request) {
	devices, err := s.store.List()
	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{
			fmt.Sprintf("error while listing store data: %s", err.Error()),
		})

		return
	}

	WriteAPIResponse(response, http.StatusOK, devices)
}

func handleStoreErr(response http.ResponseWriter, err error) {
	if errors.Is(err, persistence.ErrItemNotFound) {
		WriteErrorResponse(response, http.StatusNotFound, []string{
			http.StatusText(http.StatusNotFound),
		})

		return
	}

	WriteErrorResponse(response, http.StatusBadRequest, []string{
		http.StatusText(http.StatusBadRequest),
	})
}

func getSignedData(data string, privateKey []byte, algorithm domain.SignatureAlgorithm) ([]byte, error) {
	var (
		signer crypto.Signer
		err    error
	)

	switch algorithm {
	case domain.RSAAlgorithm:
		signer, err = crypto.NewRSASigner(privateKey)
	case domain.ECCAlgorithm:
		signer, err = crypto.NewECDSASigner(privateKey)
	default:
		return nil, fmt.Errorf("invalid algorithm %s", algorithm)
	}

	if err != nil {
		return nil, err
	}

	return signer.Sign([]byte(data))
}

func getKeys(algorithm domain.SignatureAlgorithm) ([]byte, []byte, error) {
	var (
		keyPair KeyPair
		err     error
	)

	switch algorithm {
	case domain.RSAAlgorithm:
		generator := crypto.RSAGenerator{}
		keyPair, err = generator.Generate()
	case domain.ECCAlgorithm:
		generator := crypto.ECCGenerator{}
		keyPair, err = generator.Generate()
	default:
		return nil, nil, fmt.Errorf("invalid algorithm %s", algorithm)
	}

	if err != nil {
		return nil, nil, err
	}

	return keyPair.ToBytes()
}
