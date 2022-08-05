package comms

//// SyndicateRegisterHandler request an ownership transfer of an asset
//func (s *S) SyndicateRegisterHandler(req SyndicateCreateReq, resp *SyndicateCreateResp) error {
//	serviceID, err := IsServerClient(req.ApiKey)
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("failed to get service id - AssetTransferOwnershipHandler")
//		return err
//	}
//
//	founder, err := users.UUID(uuid.FromStringOrNil(req.FoundedByID))
//	if err != nil {
//		return err
//	}
//
//	if !founder.FactionID.Valid {
//		return fmt.Errorf("user does not have faction")
//	}
//
//	isLocked := founder.CheckUserIsLocked("account")
//	if isLocked {
//		return terror.Error(fmt.Errorf("user: %s attempting to purchase on Supremacy while locked", founder.ID), "This account is locked, contact support to unlock.")
//	}
//
//	syndicateRegisterFee := db.GetDecimalWithDefault(db.KeySyndicateRegisterFee, decimal.New(5000, 18))
//	syndicateRegisterFeeCut := db.GetDecimalWithDefault(db.KeySyndicateRegisterFeeCut, decimal.NewFromFloat(0.5))
//
//	// calculate sups to new syndicate account
//	supsToSyndicateAcc := syndicateRegisterFee.Sub(syndicateRegisterFee.Mul(syndicateRegisterFeeCut))
//
//	tx, err := passdb.StdConn.Begin()
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("Failed to begin db transaction")
//		return terror.Error(err, "Failed to register syndicate in Xsyn")
//	}
//
//	defer tx.Rollback()
//
//	// create an account for the syndicate
//	account := boiler.Account{
//		ID:   req.SyndicateID,
//		Type: boiler.AccountTypeSYNDICATE,
//		Sups: decimal.Zero,
//	}
//
//	err = account.Insert(tx, boil.Infer())
//	if err != nil {
//		passlog.L.Error().Err(err).Interface("account", account).Msg("Failed to create syndicate account.")
//		return terror.Error(err, "Failed to create syndicate account")
//	}
//
//	// create syndicate
//	syndicate := boiler.Syndicate{
//		ID:          req.SyndicateID,
//		FoundedByID: founder.ID,
//		FactionID:   founder.FactionID.String,
//		Name:        req.Name,
//		AccountID:   account.ID,
//	}
//
//	err = syndicate.Insert(tx, boil.Infer())
//	if err != nil {
//		passlog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to insert syndicate into db")
//		return terror.Error(err, "Failed to register syndicate in Xsyn")
//	}
//
//	syndicateCreateTx := &types.NewTransaction{
//		Debit:                req.FoundedByID,
//		Credit:               types.SupremacyGameUserID.String(),
//		TransactionReference: types.TransactionReference(fmt.Sprintf("syndicate_create|SUPREMACY|%s|%d", req.SyndicateID, time.Now().UnixNano())),
//		Description:          "Start a new syndicate",
//		Amount:               syndicateRegisterFee,
//		Group:                types.TransactionGroupSupremacy,
//		SubGroup:             "syndicate create",
//		ServiceID:            types.UserID(uuid.FromStringOrNil(serviceID)),
//	}
//
//	_, err = s.UserCacheMap.Transact(syndicateCreateTx)
//	if err != nil {
//		return terror.Error(err, err.Error())
//	}
//
//	syndicateStartFund := &types.NewTransaction{
//		Debit:                types.SupremacyGameUserID.String(),
//		Credit:               syndicate.ID,
//		TransactionReference: types.TransactionReference(fmt.Sprintf("syndicate_start_fund|SUPREMACY|%s|%d", req.SyndicateID, time.Now().UnixNano())),
//		Description:          "Fund for starting syndicate",
//		Amount:               supsToSyndicateAcc,
//		Group:                types.TransactionGroupSupremacy,
//		SubGroup:             "syndicate create",
//		ServiceID:            types.UserID(uuid.FromStringOrNil(serviceID)),
//	}
//
//	_, err = s.UserCacheMap.Transact(syndicateStartFund)
//	if err != nil {
//		return terror.Error(err, err.Error())
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("Failed to commit db transaction")
//		return terror.Error(err, "Failed to register syndicate in Xsyn")
//	}
//
//	return nil
//}
//
//func (s *S) SyndicateNameChangeHandler(req SyndicateNameCreateReq, resp *SyndicateNameChangeResp) error {
//	_, err := IsServerClient(req.ApiKey)
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("failed to get service id - AssetTransferOwnershipHandler")
//		return err
//	}
//
//	syndicate, err := boiler.FindSyndicate(passdb.StdConn, req.SyndicateID)
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("Failed to get syndicate")
//		return terror.Error(err, "Syndicate does not exist or it is liquidated.")
//	}
//
//	syndicate.Name = req.Name
//	_, err = syndicate.Update(passdb.StdConn, boil.Whitelist(boiler.SyndicateColumns.Name))
//	if err != nil {
//		passlog.L.Error().Str("syndicate id", syndicate.ID).Str("name", req.Name).Err(err).Msg("Failed to update syndicate name")
//		return terror.Error(err, "Failed to change syndicate name")
//	}
//
//	return nil
//}
//
//func (s *S) SyndicateLiquidateHandler(req SyndicateLiquidateReq, resp *SyndicateLiquidateResp) error {
//	_, err := IsServerClient(req.ApiKey)
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("failed to get service id - AssetTransferOwnershipHandler")
//		return err
//	}
//
//	syndicate, err := boiler.FindSyndicate(passdb.StdConn, req.SyndicateID)
//	if err != nil {
//		passlog.L.Error().Err(err).Msg("Failed to get syndicate")
//		return terror.Error(err, "Syndicate does not exist or it is liquidated.")
//	}
//
//	syndicate.DeletedAt = null.TimeFrom(time.Now())
//	_, err = syndicate.Update(passdb.StdConn, boil.Whitelist(boiler.SyndicateColumns.DeletedAt))
//	if err != nil {
//		passlog.L.Error().Err(err).Str("syndicate id", syndicate.ID).Msg("Failed to archive syndicate")
//		return terror.Error(err, "Failed to liquidate syndicate")
//	}
//
//	return nil
//}
