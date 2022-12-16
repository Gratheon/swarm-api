package model

import (
	"crypto/md5"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"gitlab.com/gratheon/swarm-api/logger"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type File struct {
	Db        *sqlx.DB
	UserID    string  `db:"user_id"`
	ID        *string `json:"id" db:"id"`
	Hash      string  `db:"hash"`
	Extension string  `db:"ext"`
	URL       string  `json:"url"`
}

func (r *File) SetUp() {
	var schema = strings.Replace(
		`CREATE TABLE IF NOT EXISTS 'files' (
  'id' int unsigned NOT NULL AUTO_INCREMENT,
  'user_id' int unsigned NOT NULL,
  'hash' varchar(32) NOT NULL,
  'ext' varchar(5) NOT NULL,
  PRIMARY KEY ('id')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`, "'", "`", -1)

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	r.Db.MustExec(schema)
}

func (r *File) GetURL() string {
	return strings.Join(
		[]string{
			viper.GetString("files_base_url"),
			r.UserID,
			"/",
			r.Hash,
			".",
			r.Extension,
		}, "")
}

func (r *File) Get(id *string) (*File, error) {
	file := File{}
	err := r.Db.Get(&file, `SELECT * 
		FROM files 
		WHERE id=? AND user_id=? 
		LIMIT 1`, id, r.UserID)
	logger.LogInfo("file")
	logger.LogInfo(file)

	file.URL = file.GetURL()
	logger.LogInfo(file)
	return &file, err
}

func (r *File) Save(content []byte) (*string, error) {
	md5 := md5.Sum(content)
	hash := fmt.Sprintf("%x", md5)
	filepath := strings.Join([]string{
		"public/files/",
		hash,
		".jpg",
	}, "")

	err := ioutil.WriteFile(filepath, content, 0644)

	if err != nil {
		return nil, err
	}

	r.upload(strings.Join([]string{
		r.UserID,
		"//",
		hash,
		".jpg",
	}, ""), filepath)

	result, err := r.Db.NamedExec(
		`INSERT INTO files (
		  user_id,
		  hash,
		  ext
		) VALUES (
		    :userID,
		  	:hash,
		    'jpg'
		)`,
		map[string]interface{}{
			"userID": r.UserID,
			"hash":   hash,
		},
	)

	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	id, err2 := result.LastInsertId()
	sid := strconv.FormatInt(id, 10)

	err3 := os.Remove(filepath)

	if err3 != nil {
		logger.LogError(err3)
		return nil, err3
	}

	return &sid, err2
}

func (r *File) upload(filename string, filepath string) {
	bucket := viper.GetString("aws_bucket")
	aws_key := viper.GetString("aws_key")
	aws_secret := viper.GetString("aws_secret")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
		Credentials: credentials.NewStaticCredentials(
			aws_key,
			aws_secret,
			"",
		),
	})
	uploader := s3manager.NewUploader(sess)

	file, err := os.Open(filepath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}

	defer file.Close()

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	logger.LogInfo(result)

	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}
