/**
 * This file was generated by Jabuti Transformation Engine.
 *
 * Copyright: Applied Computing Research Group (GCA), Unijuí University, Ijui-RS, Brazil
 * SPDX-License-Identifier: MIT
 * Author: Mailson Teles Borges <mailson.borges@sou.unijui.edu.br>
 *
 * For more information about the Jabuti Transformation Engine, please visit:
 * https://github.com/gca-research-group/jabuti-ce-transformation-engine
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Party struct {
	Id            string
	Name          string
	IsSigned      bool
	SignatureDate time.Time
}

type Parties struct {
	Application Party
	Process     Party
}

type Interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Timeout struct {
	Increase int       `json:"increase"`
	End      time.Time `json:"end"`
}

type MaxNumberOfOperation struct {
	Max      int       `json:"max"`
	Used     int       `json:"used"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	TimeUnit string    `json:"timeUnit"`
}

type RightRequestDelivery struct {
}

type RightRequestDeliveryArgs struct {
	NumberOfAddresses int `json:"numberOfAddresses"`

	Weight int `json:"weight"`

	ProductValue int `json:"productValue"`
}

type Request struct {
	clientId  string
	createdAt time.Time
}

type Asset struct {
	Parties   Parties
	BeginDate time.Time
	DueDate   time.Time
	IsSigned  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Requests  map[string]Request

	RightRequestDelivery RightRequestDelivery
}

type PartyRequest struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type PartiesRequest struct {
	Application PartyRequest `json:"application"`
	Process     PartyRequest `json:"process"`
}

type AssetRequest struct {
	BeginDate string         `json:"beginDate"`
	DueDate   string         `json:"dueDate"`
	Parties   PartiesRequest `json:"parties"`
}

func (s *SmartContract) isParty(id string, asset *Asset) (bool, error) {
	value := id == asset.Parties.Process.Id || id == asset.Parties.Application.Id

	if !value {
		return value, fmt.Errorf("only the process or the application can execute this operation")
	}

	return value, nil
}

func (s *SmartContract) isSigned(party Party) (bool, error) {
	if party.IsSigned {
		return party.IsSigned, fmt.Errorf("the asset is already signed")
	}

	return party.IsSigned, nil
}

func (s *SmartContract) assetIsSigned(asset *Asset) error {
	if asset.IsSigned {
		return nil
	}

	return fmt.Errorf("asset is not signed")
}

func (s *SmartContract) isBetweenBeginDateAndDueDate(asset *Asset) error {
	if asset.DueDate.Before(time.Now()) {
		return fmt.Errorf("asset expired. The current date is after the due date")
	}

	if asset.BeginDate.After(time.Now()) {
		return fmt.Errorf("the current date is before the start date")
	}

	return nil
}

func (s *SmartContract) isApplicationIdValid(id string) error {
	if id == "" {
		return fmt.Errorf("application id is required")
	}

	return nil
}

func (s *SmartContract) isProcessIdValid(id string) error {
	if id == "" {
		return fmt.Errorf("process id is required")
	}

	return nil
}

func (s *SmartContract) isBeginDateValid(beginDate time.Time) error {
	if beginDate.IsZero() {
		return fmt.Errorf("begin date is required")
	}

	return nil
}

func (s *SmartContract) isDueDateValid(dueDate time.Time) error {
	if dueDate.IsZero() {
		return fmt.Errorf("due date is required")
	}

	return nil
}

func (s *SmartContract) isDueDateGreaterThanBeginDate(beginDate time.Time, dueDate time.Time) error {
	if beginDate.After(dueDate) {
		return fmt.Errorf("begin date greater than due date")
	}

	return nil
}

func (s *SmartContract) string2Time(date string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, date)

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date. Expected format 2006-01-02T15:04:05Z07:00. Recieved: %s", err.Error())
	}

	return parsed, nil
}

func (s *SmartContract) putState(ctx contractapi.TransactionContextInterface, assetId string, asset *Asset) error {
	contractAsBytes, err := json.Marshal(asset)

	if err != nil {
		return fmt.Errorf("marshal error: %s", err.Error())
	}

	ctx.GetStub().PutState(assetId, contractAsBytes)

	return nil
}

func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface, assetRequest AssetRequest) (string, error) {
	var beginDate time.Time
	var dueDate time.Time
	var err error

	if beginDate, err = s.string2Time(assetRequest.BeginDate); err != nil {
		return "", err
	}

	if dueDate, err = s.string2Time(assetRequest.DueDate); err != nil {
		return "", err
	}

	if err := s.isBeginDateValid(beginDate); err != nil {
		return "", err
	}

	if err := s.isDueDateValid(dueDate); err != nil {
		return "", err
	}

	if err := s.isDueDateGreaterThanBeginDate(beginDate, dueDate); err != nil {
		return "", err
	}

	if err := s.isApplicationIdValid(assetRequest.Parties.Process.Id); err != nil {
		return "", err
	}

	if err := s.isProcessIdValid(assetRequest.Parties.Process.Id); err != nil {
		return "", err
	}

	asset := Asset{}
	parties := Parties{}

	asset.BeginDate = beginDate
	asset.DueDate = dueDate

	parties.Application.Id = assetRequest.Parties.Application.Id
	parties.Application.Name = assetRequest.Parties.Application.Name
	parties.Application.IsSigned = false

	parties.Process.Id = assetRequest.Parties.Process.Id
	parties.Process.Name = assetRequest.Parties.Process.Name
	parties.Process.IsSigned = false

	asset.Parties = parties
	asset.CreatedAt = time.Now()
	asset.Requests = make(map[string]Request)

	assetId := uuid.New().String()

	s.putState(ctx, assetId, &asset)

	return assetId, nil
}

func (s *SmartContract) Sign(ctx contractapi.TransactionContextInterface, assetId string) error {

	var id string
	var err error
	var asset *Asset

	if id, err = s.QueryClientId(ctx); err != nil {
		return err
	}

	if asset, err = s.QueryAsset(ctx, assetId); err != nil {
		return err
	}

	if _, err := s.isParty(id, asset); err != nil {
		return err
	}

	if err := s.isBetweenBeginDateAndDueDate(asset); err != nil {
		return err
	}

	if asset.Parties.Application.Id == id {

		if _, err := s.isSigned(asset.Parties.Application); err != nil {
			return err
		}

		asset.Parties.Application.IsSigned = true
		asset.Parties.Application.SignatureDate = time.Now()
	}

	if asset.Parties.Process.Id == id {

		if _, err := s.isSigned(asset.Parties.Process); err != nil {
			return err
		}

		asset.Parties.Process.IsSigned = true
		asset.Parties.Process.SignatureDate = time.Now()
	}

	asset.IsSigned = asset.Parties.Application.IsSigned && asset.Parties.Process.IsSigned

	s.putState(ctx, assetId, asset)

	return nil
}

func (s *SmartContract) QueryAsset(ctx contractapi.TransactionContextInterface, assetId string) (*Asset, error) {

	contractAsBytes, err := ctx.GetStub().GetState(assetId)

	if err != nil {
		return nil, fmt.Errorf("failed to read from state: %s", err.Error())
	}

	if contractAsBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", assetId)
	}

	asset := new(Asset)

	err = json.Unmarshal(contractAsBytes, asset)

	if err != nil {
		return nil, fmt.Errorf("marshal error: %s", err.Error())
	}

	return asset, nil
}

func (s *SmartContract) QueryClientId(ctx contractapi.TransactionContextInterface) (string, error) {
	clientIdentity := ctx.GetClientIdentity()

	if clientIdentity == nil {
		return "", fmt.Errorf("failed to get client identity")
	}

	clientID, err := clientIdentity.GetID()

	if err != nil {
		return "", fmt.Errorf("failed to get client ID: %v", err)
	}

	return clientID, nil
}

func (s *SmartContract) ClauseRightRequestDelivery(ctx contractapi.TransactionContextInterface, assetId string, args RightRequestDeliveryArgs) (bool, error) {

	var err error
	var asset *Asset

	if asset, err = s.QueryAsset(ctx, assetId); err != nil {
		return false, err
	}

	if err = s.isBetweenBeginDateAndDueDate(asset); err != nil {
		return false, err
	}

	if err = s.assetIsSigned(asset); err != nil {
		return false, err
	}

	id := uuid.New().String()
	createdAt := time.Now()

	var clientId string

	if clientId, err = s.QueryClientId(ctx); err != nil {
		return false, err
	}

	asset.Requests[id] = Request{
		clientId:  clientId,
		createdAt: createdAt,
	}

	s.putState(ctx, assetId, asset)

	isValid := true

	isValid = isValid && args.NumberOfAddresses == 1

	isValid = isValid && args.Weight == 100

	isValid = isValid && args.ProductValue < 20000

	if !isValid {
		return isValid, fmt.Errorf("Request operation did not meet all requirements")
	}

	return isValid, nil
}

func main() {
	chainconde, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		log.Panicf("error create chaincode: %s", err.Error())
		return
	}

	if err := chainconde.Start(); err != nil {
		log.Panicf("error create chaincode: %s", err.Error())
	}
}
