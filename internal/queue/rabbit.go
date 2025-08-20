// internal/queue/rabbit.go
//
// بسته‌بندی ساده RabbitMQ: اتصال و کانال را نگه می‌دارد، صف‌های مورد نیاز را اعلام می‌کند،
// و کمک‌کننده‌هایی برای انتشار پیام‌های JSON/متن ارائه می‌دهد.

package queue

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/streadway/amqp"

	"github.com/iliyamo/Library-Management-System/internal/model"
)

// GetRabbitClient کلاینت RabbitMQ را برای دسترسی خارجی برمی‌گرداند.
func GetRabbitClient() *RabbitMQClient {
	return rabbitClient // rabbitClient که در همین فایل تعریف شده رو برمی‌گردونه
}

// RabbitMQClient اتصال و کانال AMQP را برای انتشار پیام‌ها نگه می‌دارد.
type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// متغیرهای سطح پکیج (singletonها)
var (
	rabbitClient  *RabbitMQClient
	rabbitLogger  *log.Logger
	rabbitLogFile *os.File
)

// InitRabbitMQ اتصال به RabbitMQ را برقرار می‌کند و صف‌های پایدار را اعلام می‌کند.
// اگر قبلاً اولیه‌سازی شده باشد، هیچ کاری نمی‌کند.
func InitRabbitMQ(amqpURL string) error {
	if rabbitClient != nil {
		return nil
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("اتصال به RabbitMQ شکست خورد: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("باز کردن کانال شکست خورد: %w", err)
	}

	// اعلام صف‌های مورد نیاز (پایدار)
	if err := declareQueues(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}

	rabbitClient = &RabbitMQClient{Conn: conn, Channel: ch}

	// لاگر اختصاصی → stdout + فایل
	if rabbitLogger == nil {
		if f, errf := os.OpenFile("rabbitmq.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); errf == nil {
			rabbitLogFile = f
			multi := io.MultiWriter(os.Stdout, f)
			rabbitLogger = log.New(multi, "[RabbitMQ] ", log.LstdFlags)
		} else {
			rabbitLogger = log.New(os.Stdout, "[RabbitMQ] ", log.LstdFlags)
		}
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("✅ RabbitMQ متصل شد؛ صف‌ها اعلام شدند: %q, %q", LoanEventsQueue, LoanCommandsQueue)
	} else {
		log.Printf("✅ RabbitMQ متصل شد؛ صف‌ها اعلام شدند: %q, %q", LoanEventsQueue, LoanCommandsQueue)
	}
	return nil
}

// declareQueues مطمئن می‌شود صف‌های مورد نیاز وجود دارند (پایدار، بدون حذف خودکار).
func declareQueues(ch *amqp.Channel) error {
	queues := []string{LoanEventsQueue, LoanCommandsQueue}
	for _, q := range queues {
		if _, err := ch.QueueDeclare(
			q,
			true,  // پایدار
			false, // حذف خودکار
			false, // انحصاری
			false, // بدون انتظار
			nil,   // آرگومان‌ها
		); err != nil {
			return fmt.Errorf("اعلام صف %q شکست خورد: %w", q, err)
		}
	}
	return nil
}

// PublishToRabbit بایت‌های خام را به صف خاصی منتشر می‌کند (تبادل پیش‌فرض مستقیم).
func PublishToRabbit(queue string, body []byte, contentType string) error {
	if rabbitClient == nil || rabbitClient.Channel == nil {
		return fmt.Errorf("RabbitMQ اولیه‌سازی نشده است")
	}
	if contentType == "" {
		contentType = "text/plain"
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("انتشار به %s (%d بایت)", queue, len(body))
	}
	return rabbitClient.Channel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		},
	)
}

// PublishLoanEvent یک LoanEvent را marshal می‌کند و به صف loan_events منتشر می‌کند.
func PublishLoanEvent(event model.LoanEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal رویداد شکست خورد: %w", err)
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("انتشار رویداد به %s: %+v", LoanEventsQueue, event)
	}
	return PublishToRabbit(LoanEventsQueue, b, "application/json")
}

// CloseRabbitMQ کانال/اتصال و فایل لاگ اختصاصی را می‌بندد.
func CloseRabbitMQ() {
	if rabbitClient != nil {
		if rabbitClient.Channel != nil {
			_ = rabbitClient.Channel.Close()
		}
		if rabbitClient.Conn != nil {
			_ = rabbitClient.Conn.Close()
		}
		rabbitClient = nil
	}
	if rabbitLogFile != nil {
		_ = rabbitLogFile.Close()
		rabbitLogFile = nil
	}
	rabbitLogger = nil
}
