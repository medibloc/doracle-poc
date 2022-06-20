package subscribe

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func CmdSubscribe() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe event",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := cmd.Flags().GetString("websocketAddr")
			if err != nil {
				return nil
			}

			event, err := cmd.Flags().GetString("event")
			if err != nil {
				return nil
			}

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket"}
			log.Printf("connecting to %s", u.String())

			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			defer c.Close()

			query := fmt.Sprintf(`{ "jsonrpc": "2.0", "method": "subscribe", "params": ["tm.event='%s'"], "id": 1 }`, event)

			err = c.WriteMessage(websocket.TextMessage, []byte(query))
			if err != nil {
				log.Fatal("dial:", err)
			}

			done := make(chan struct{})
			readMessage(done, c)

			for {
				select {
				case <-done:
					return nil
				case <-interrupt:
					log.Println("interrupt")
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Println("write close:", err)
						return err
					}
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					return nil
				}
			}
		},
	}

	cmd.Flags().String("websocketAddr", "", "blockChain websocket address")
	cmd.Flags().String("event", "NewBlock", "'NewBlock', 'NewBlockHeader', 'NewEvidence', 'Tx', 'ValidatorSetUpdates'")
	return cmd
}

func readMessage(done chan struct{}, c *websocket.Conn) {
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			log.Println(string(message))

		}
	}()
}
