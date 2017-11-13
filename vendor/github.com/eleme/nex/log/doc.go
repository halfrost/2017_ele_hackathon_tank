// Package log is wrapper built on top of https://github.com/eleme/log.
//
// If you use nex as framework to built your own app, please do not use
// your own logging implementation, the usage is simple like that:
//
//   logger, err := log.GetContextLogger("whatever") // log.GetLogger is not recommended for use
//   if err != nil {
//       // DO NOT ignore errors.
//   }
//   logger.ContextInfof(context.TODO(), "hello %v", "world")
//
// About the Sensitive Argument Fields in Server Side Log.
//
// You should find out the some xxxRequest structs at the end of handler/auto-endpoints.go,
// for example, we have a handler and a request struct:
//   // // Thrift struct
//   // type User struct {
//   //     Name string       `...`
//   //     Password string   `...`
//   //     Phone string      `...`
//   // }
//   (h *XXXHandler) Register(ctx context.Context, whatever string, user *User) (bool, error) {
//       // do sth.
//   }
//
// We will generate a struct named registerRequest:
//   type registerRequest struct {
//       Whatever string
//       User     *User
//   }
//
// As you can see, the struct's fields is corresponds to the handler's arguments.
// If you want to hide the Password and Phone field, you can implements the ALogStringer
// interface, simple like that:
//   func (req registerRequest) ALogString() string {
//       return fmt.Sprintf("%v, {username: %v, password: ***, phone: ***}", req.Whatever, req.User.Name)
//   }
//
package log
