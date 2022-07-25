package api

//type IsLocked struct {
//	deadlock.RWMutex
//	isLocked bool
//}

//type TransactionCache struct {
//	deadlock.RWMutex
//	conn         *sql.DB
//	log          *zerolog.Logger
//	transactions []*types.NewTransaction
//	IsLocked     *IsLocked
//}

//func NewTransactionCache(conn *sql.DB, log *zerolog.Logger) *TransactionCache {
//	tc := &TransactionCache{
//		deadlock.RWMutex{},
//		conn,
//		log,
//		[]*types.NewTransaction{},
//		&IsLocked{
//			isLocked: false,
//		},
//	}
//
//	ticker := time.NewTicker(10 * time.Second)
//
//	go func() {
//		for {
//			<-ticker.C
//			tc.commit()
//		}
//	}()
//
//	return tc
//}

//func (tc *TransactionCache) commit() {
//	tc.Lock()
//	ctrans := make([]*types.NewTransaction, len(tc.transactions))
//	copy(ctrans, tc.transactions)
//	tc.transactions = []*types.NewTransaction{}
//	tc.Unlock()
//	for _, tx := range ctrans {
//		err := CreateTransactionEntry(
//			tc.conn,
//			tx,
//		)
//
//		if err != nil {
//			if tx.NotSafe {
//				tc.
//					log.
//					Err(err).
//					Str("amt", tx.Amount.String()).
//					Str("from", tx.From.String()).
//					Str("to", tx.To.String()).
//					Str("txref", string(tx.TransactionReference)).
//					Msg("transaction cache lock")
//				tc.IsLocked.Lock()
//				tc.IsLocked.isLocked = true
//				tc.IsLocked.Unlock()
//				tc.Lock() //grind to a halt if transactions fail to save to database
//			}
//			return
//		}
//	}
//}

//func (tc *TransactionCache) Process(t *types.NewTransaction) string {
//
//	if t.Processed {
//		return t.ID
//	}
//	t.ID = fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
//	t.CreatedAt = time.Now()
//	t.Processed = true
//
//	tc.Lock()
//	defer func() {
//		tc.Unlock()
//		if !t.NotSafe {
//			tc.commit()
//		}
//	}()
//	tc.transactions = append(tc.transactions, t)
//
//	return t.ID
//}
