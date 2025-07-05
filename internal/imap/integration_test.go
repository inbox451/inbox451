package imap

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"inbox451/internal/config"
	"inbox451/internal/core"
	"inbox451/internal/logger"
	"inbox451/internal/models"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	null "github.com/volatiletech/null/v9"
)

// IMAPIntegrationTestSuite provides integration testing for the IMAP server
type IMAPIntegrationTestSuite struct {
	suite.Suite
	core       *core.Core
	imapServer *ImapServer
	client     *client.Client
	testPort   string

	// Test data
	testUser    *models.User
	testProject *models.Project
	testInbox   *models.Inbox
	testMessage *models.Message
	testToken   *models.Token
}

func TestIMAPIntegrationSuite(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IMAPIntegrationTestSuite))
}

func (suite *IMAPIntegrationTestSuite) SetupSuite() {
	var err error

	// Load test configuration using koanf
	ko := koanf.New(".")
	cfg, err := config.LoadConfig("../../config.yml", ko)
	if err != nil {
		// Fallback to default test configuration
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				URL:             "postgres://inbox:inbox@localhost:5432/inbox451?sslmode=disable",
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
		}
		cfg.Logging.Level = logger.DEBUG
	}

	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	require.NoError(suite.T(), err, "Failed to connect to test database")

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Initialize core with proper parameters
	suite.core, err = core.NewCore(cfg, db, "test", "test", "test")
	require.NoError(suite.T(), err, "Failed to initialize core")

	// Create IMAP server
	suite.imapServer, err = NewServer(suite.core)
	require.NoError(suite.T(), err, "Failed to create IMAP server")

	// Set the test port based on the server's configured address
	serverAddr := suite.imapServer.imap.Addr
	if serverAddr == ":1143" || serverAddr == "" {
		suite.testPort = "localhost:1143"
	} else {
		suite.testPort = serverAddr
	}

	// Start IMAP server in goroutine
	go func() {
		err := suite.imapServer.ListenAndServe()
		if err != nil {
			log.Printf("IMAP server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Set up test data
	suite.setupTestData()
}

func (suite *IMAPIntegrationTestSuite) TearDownSuite() {
	if suite.client != nil {
		err := suite.client.Close()
		if err != nil {
			return
		}
	}

	if suite.imapServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := suite.imapServer.Shutdown(ctx); err != nil {
			suite.T().Logf("Failed to shutdown IMAP server: %v", err)
		}
	}

	// Clean up test data
	suite.cleanupTestData()
}

func (suite *IMAPIntegrationTestSuite) SetupTest() {
	var err error

	// Connect to IMAP server
	suite.client, err = client.Dial(suite.testPort)
	require.NoError(suite.T(), err, "Failed to connect to IMAP server")

	// Ensure we have a fresh test message for each test
	suite.ensureTestMessage()
}

func (suite *IMAPIntegrationTestSuite) ensureTestMessage() {
	ctx := context.Background()

	// Always create a fresh unread message for each test
	// First, delete any existing messages to ensure clean state
	isDeleted := false
	filters := models.MessageFilters{
		IsDeleted: &isDeleted,
	}
	response, err := suite.core.MessageService.ListByInbox(ctx, suite.testInbox.ID, 100, 0, filters)
	if err == nil && response != nil {
		messages := response.Data.([]*models.Message)
		for _, msg := range messages {
			suite.core.MessageService.Delete(ctx, msg.ID)
		}
	}

	// Create a fresh test message
	message := &models.Message{
		InboxID:   suite.testInbox.ID,
		Sender:    "sender@example.com",
		Receiver:  suite.testInbox.Email,
		Subject:   "Test Message",
		Body:      "This is a test message body",
		IsRead:    false,
		IsDeleted: false,
	}

	err = suite.core.MessageService.Store(ctx, message)
	require.NoError(suite.T(), err, "Failed to create fresh test message")
}

func (suite *IMAPIntegrationTestSuite) TearDownTest() {
	if suite.client != nil {
		if err := suite.client.Logout(); err != nil {
			suite.T().Logf("Failed to logout: %v", err)
		}
		err := suite.client.Close()
		if err != nil {
			return
		}
		suite.client = nil
	}
}

func (suite *IMAPIntegrationTestSuite) setupTestData() {
	ctx := context.Background()

	// Generate unique identifiers for this test run
	timestamp := time.Now().Unix()

	// Create test user with unique identifiers
	user := &models.User{
		Name:          fmt.Sprintf("Test User %d", timestamp),
		Username:      fmt.Sprintf("testuser%d", timestamp),
		Email:         fmt.Sprintf("testuser%d@example.com", timestamp),
		Status:        "active",
		Role:          "user",
		PasswordLogin: true,
		Password:      null.StringFrom("password123"), // Will be hashed by UserService.Create
	}

	var err error
	err = suite.core.UserService.Create(ctx, user)
	require.NoError(suite.T(), err, "Failed to create test user")
	suite.testUser = user

	// Create test project with unique name
	project := &models.Project{
		Name: fmt.Sprintf("Test Project %d", timestamp),
	}

	err = suite.core.ProjectService.Create(ctx, project)
	require.NoError(suite.T(), err, "Failed to create test project")
	suite.testProject = project

	// Add user to project
	projectUser := &models.ProjectUser{
		ProjectID: suite.testProject.ID,
		UserID:    suite.testUser.ID,
		Role:      "admin",
	}
	err = suite.core.ProjectService.AddUser(ctx, projectUser)
	require.NoError(suite.T(), err, "Failed to add user to project")

	// Create test inbox with unique email
	inboxEmail := fmt.Sprintf("testinbox%d@example.com", timestamp)
	inbox := &models.Inbox{
		ProjectID: suite.testProject.ID,
		Email:     inboxEmail,
	}

	err = suite.core.InboxService.Create(ctx, inbox)
	require.NoError(suite.T(), err, "Failed to create test inbox")
	suite.testInbox = inbox

	// Create test message
	message := &models.Message{
		InboxID:   suite.testInbox.ID,
		Sender:    "sender@example.com",
		Receiver:  inboxEmail,
		Subject:   "Test Message",
		Body:      "This is a test message body",
		IsRead:    false,
		IsDeleted: false,
	}

	err = suite.core.MessageService.Store(ctx, message)
	require.NoError(suite.T(), err, "Failed to create test message")
	suite.testMessage = message

	// Create API token for test user
	tokenData := &models.Token{
		Name: "Test Token",
	}
	token, err := suite.core.TokenService.CreateForUser(ctx, suite.testUser.ID, tokenData)
	require.NoError(suite.T(), err, "Failed to create test token")
	suite.testToken = token
}

func (suite *IMAPIntegrationTestSuite) cleanupTestData() {
	ctx := context.Background()

	if suite.testMessage != nil {
		if err := suite.core.MessageService.Delete(ctx, suite.testMessage.ID); err != nil {
			suite.T().Logf("Failed to delete test message: %v", err)
		}
	}

	if suite.testInbox != nil {
		if err := suite.core.InboxService.Delete(ctx, suite.testInbox.ID); err != nil {
			suite.T().Logf("Failed to delete test inbox: %v", err)
		}
	}

	if suite.testProject != nil {
		if err := suite.core.ProjectService.Delete(ctx, suite.testProject.ID); err != nil {
			suite.T().Logf("Failed to delete test project: %v", err)
		}
	}

	if suite.testUser != nil {
		if err := suite.core.UserService.Delete(ctx, suite.testUser.ID); err != nil {
			suite.T().Logf("Failed to delete test user: %v", err)
		}
	}
}

func (suite *IMAPIntegrationTestSuite) TestAuthentication() {
	suite.T().Run("ValidCredentials", func(t *testing.T) {
		err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
		assert.NoError(t, err, "Login with valid credentials should succeed")
	})

	suite.T().Run("InvalidCredentials", func(t *testing.T) {
		// Create new client for this test to avoid state issues
		testClient, err := client.Dial(suite.testPort)
		require.NoError(t, err)
		defer func(testClient *client.Client) {
			err := testClient.Close()
			if err != nil {
				suite.T().Logf("Failed to close test client: %v", err)
			}
		}(testClient)

		err = testClient.Login(suite.testUser.Username, "wrongtoken")
		assert.Error(t, err, "Login with invalid credentials should fail")
	})

	suite.T().Run("NonexistentUser", func(t *testing.T) {
		// Create new client for this test
		testClient, err := client.Dial(suite.testPort)
		require.NoError(t, err)
		defer func(testClient *client.Client) {
			err := testClient.Close()
			if err != nil {
				suite.T().Logf("Failed to close test client: %v", err)
			}
		}(testClient)

		err = testClient.Login("nonexistent", "invalidtoken")
		assert.Error(t, err, "Login with nonexistent user should fail")
	})
}

func (suite *IMAPIntegrationTestSuite) TestMailboxOperations() {
	// First login
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	suite.T().Run("ListMailboxes", func(t *testing.T) {
		mailboxes := make(chan *imap.MailboxInfo, 10)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.List("", "*", mailboxes)
		}()

		var foundInboxes []*imap.MailboxInfo
		for mailbox := range mailboxes {
			foundInboxes = append(foundInboxes, mailbox)
		}

		err := <-done
		assert.NoError(t, err, "LIST command should succeed")

		// Should find at least our test inbox
		assert.True(t, len(foundInboxes) > 0, "Should find at least one mailbox")

		// Check if our test inbox is in the list
		found := false
		for _, mailbox := range foundInboxes {
			if mailbox.Name == suite.testInbox.Email {
				found = true
				break
			}
		}
		assert.True(t, found, "Test inbox should be in the mailbox list")
	})

	suite.T().Run("SelectMailbox", func(t *testing.T) {
		mbox, err := suite.client.Select(suite.testInbox.Email, false)
		assert.NoError(t, err, "SELECT command should succeed")
		assert.NotNil(t, mbox, "Mailbox status should not be nil")
		assert.Equal(t, suite.testInbox.Email, mbox.Name, "Selected mailbox name should match")
		assert.True(t, mbox.Messages > 0, "Mailbox should contain at least one message")
	})
}

func (suite *IMAPIntegrationTestSuite) TestMessageFetching() {
	// Login and select mailbox
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	_, err = suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	suite.T().Run("FetchBasicInfo", func(t *testing.T) {
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 0) // All messages

		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid}, messages)
		}()

		var fetchedMessages []*imap.Message
		for msg := range messages {
			fetchedMessages = append(fetchedMessages, msg)
		}

		err := <-done
		assert.NoError(t, err, "FETCH command should succeed")
		assert.True(t, len(fetchedMessages) > 0, "Should fetch at least one message")

		// Check first message
		msg := fetchedMessages[0]
		assert.NotNil(t, msg.Envelope, "Message envelope should not be nil")
		assert.Equal(t, suite.testMessage.Subject, msg.Envelope.Subject, "Subject should match")
	})

	suite.T().Run("FetchBody", func(t *testing.T) {
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 1) // First message only

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)

		section := &imap.BodySectionName{}

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
		}()

		var msg *imap.Message
		for m := range messages {
			msg = m
		}

		err := <-done
		assert.NoError(t, err, "FETCH BODY command should succeed")
		assert.NotNil(t, msg, "Should receive a message")

		body := msg.GetBody(section)
		assert.NotNil(t, body, "Message body should not be nil")
	})
}

func (suite *IMAPIntegrationTestSuite) TestMessageSearch() {
	// Login and select mailbox
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	_, err = suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	suite.T().Run("SearchAll", func(t *testing.T) {
		criteria := imap.NewSearchCriteria()

		uids, err := suite.client.Search(criteria)
		assert.NoError(t, err, "SEARCH ALL should succeed")
		assert.True(t, len(uids) > 0, "Should find at least one message")
	})

	suite.T().Run("SearchUnseen", func(t *testing.T) {
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.SeenFlag}

		uids, err := suite.client.Search(criteria)
		assert.NoError(t, err, "SEARCH UNSEEN should succeed")
		// Since our test message is unread, we should find it
		assert.True(t, len(uids) > 0, "Should find unread messages")
	})
}

func (suite *IMAPIntegrationTestSuite) TestFlagOperations() {
	// Login and select mailbox
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	_, err = suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	suite.T().Run("AddSeenFlag", func(t *testing.T) {
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 1) // First message

		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []any{imap.SeenFlag}

		err := suite.client.Store(seqset, item, flags, nil)
		assert.NoError(t, err, "STORE +FLAGS \\Seen should succeed")

		// Verify the flag was set by fetching flags
		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{imap.FetchFlags}, messages)
		}()

		var msg *imap.Message
		for m := range messages {
			msg = m
		}

		err = <-done
		assert.NoError(t, err, "FETCH FLAGS should succeed")
		assert.Contains(t, msg.Flags, imap.SeenFlag, "Message should have \\Seen flag")
	})

	suite.T().Run("AddDeletedFlag", func(t *testing.T) {
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 1) // First message

		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []any{imap.DeletedFlag}

		err := suite.client.Store(seqset, item, flags, nil)
		assert.NoError(t, err, "STORE +FLAGS \\Deleted should succeed")

		// Verify the flag was set
		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{imap.FetchFlags}, messages)
		}()

		var msg *imap.Message
		for m := range messages {
			msg = m
		}

		err = <-done
		assert.NoError(t, err, "FETCH FLAGS should succeed")
		assert.Contains(t, msg.Flags, imap.DeletedFlag, "Message should have \\Deleted flag")
	})

	suite.T().Run("RemoveFlags", func(t *testing.T) {
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 1) // First message

		item := imap.FormatFlagsOp(imap.RemoveFlags, true)
		flags := []any{imap.DeletedFlag}

		err := suite.client.Store(seqset, item, flags, nil)
		assert.NoError(t, err, "STORE -FLAGS \\Deleted should succeed")

		// Verify the flag was removed
		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{imap.FetchFlags}, messages)
		}()

		var msg *imap.Message
		for m := range messages {
			msg = m
		}

		err = <-done
		assert.NoError(t, err, "FETCH FLAGS should succeed")
		assert.NotContains(t, msg.Flags, imap.DeletedFlag, "Message should not have \\Deleted flag")
	})
}

func (suite *IMAPIntegrationTestSuite) TestExpunge() {
	// Login and select mailbox
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	mbox, err := suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	initialCount := mbox.Messages

	// Mark first message as deleted
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, 1)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []any{imap.DeletedFlag}

	err = suite.client.Store(seqset, item, flags, nil)
	require.NoError(suite.T(), err)

	// Expunge
	err = suite.client.Expunge(nil)
	assert.NoError(suite.T(), err, "EXPUNGE should succeed")

	// Check that message count decreased
	mbox, err = suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	assert.Less(suite.T(), mbox.Messages, initialCount, "Message count should decrease after expunge")
}

// Benchmark tests for performance
func (suite *IMAPIntegrationTestSuite) TestPerformance() {
	// Login
	err := suite.client.Login(suite.testUser.Username, suite.testToken.Token)
	require.NoError(suite.T(), err)

	_, err = suite.client.Select(suite.testInbox.Email, false)
	require.NoError(suite.T(), err)

	suite.T().Run("BenchmarkFetch", func(t *testing.T) {
		start := time.Now()

		seqset := new(imap.SeqSet)
		seqset.AddRange(1, 0) // All messages

		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)

		go func() {
			done <- suite.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, messages)
		}()

		var count int
		for range messages {
			count++
		}

		err := <-done
		assert.NoError(t, err)

		duration := time.Since(start)
		t.Logf("Fetched %d messages in %v (%.2f msg/sec)", count, duration, float64(count)/duration.Seconds())
	})
}
