package user

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"hash"

	_ "github.com/mattn/go-sqlite3" // sqlite driver requires such import
)

func createGrafanaUser(newUser PMMUser) error {
	email := newUser.Username + "@localhost"
	salt := getRandomString(10)
	rands := getRandomString(10)
	password := encodePassword(newUser.Password, salt)

	db, err := sql.Open("sqlite3", SSMConfig.GrafanaDBPath)
	if err != nil {
		return err
	}
	defer db.Close() // nolint: errcheck

	if _, err = db.Exec("PRAGMA busy_timeout = 60000"); err != nil {
		return err
	}

	affect, err := updateUser(db, newUser.Username, email, password, salt, rands)
	if err != nil {
		return err
	}

	if affect == 0 {
		userID, err := insertUser(db, newUser.Username, email, password, salt, rands)
		if err != nil {
			return err
		}
		return addUserToOrg(db, userID)
	}

	return nil
}

func updateUser(db *sql.DB, username, email, password, salt, rands string) (int64, error) {
	stmt, err := db.Prepare(`
        UPDATE user
        SET    version  = 1,
               email    = ?,
               password = ?,
               salt     = ?,
               rands    = ?,
               updated  = date('now')
        WHERE  login    = ?
    `)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(email, password, salt, rands, username)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func insertUser(db *sql.DB, username, email, password, salt, rands string) (int64, error) {
	stmt, err := db.Prepare(`
        INSERT INTO user (version, login, email, password, salt, rands, org_id, is_admin,     created,     updated)
        VALUES           (      1,     ?,     ?,        ?,    ?,     ?,      1,        1, date('now'), date('now'))
    `)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(username, email, password, salt, rands)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func addUserToOrg(db *sql.DB, userID int64) error {
	stmt, err := db.Prepare(`
        INSERT INTO org_user (org_id, user_id,    role,     created,     updated)
        VALUES               (     1,       ?, 'Admin', date('now'), date('now'));
    `)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(userID); err != nil {
		return err
	}
	return nil
}

func deleteGrafanaUser(username string) error {
	db, err := sql.Open("sqlite3", SSMConfig.GrafanaDBPath)
	if err != nil {
		return err
	}
	defer db.Close() // nolint: errcheck

	stmt, err := db.Prepare("DELETE FROM user WHERE login = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username)
	return err
}

// source: https://github.com/grafana/grafana/blob/v5.1.3/pkg/util/encoding.go#L16
func getRandomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		if len(alphabets) == 0 {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return string(bytes)
}

// source: https://github.com/grafana/grafana/blob/v5.1.3/pkg/util/encoding.go#L30
func encodePassword(password string, salt string) string {
	newPasswd := PBKDF2([]byte(password), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(newPasswd)
}

// source: https://github.com/grafana/grafana/blob/v5.1.3/pkg/util/encoding.go#L43
func PBKDF2(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		// N.B.: || means concatenation, ^ means XOR
		// for each block T_i = U_1 ^ U_2 ^ ... ^ U_iter
		// U_1 = PRF(password, salt || uint(i))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)

		// U_n = PRF(password, U_(n-1))
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	return dk[:keyLen]
}
