package service

type buildMessage interface {
	registerAccount(codeRegister string) string
}

type msg struct{}

func (m *msg) registerAccount(codeRegister string) string {
	return "VERIFY:" + codeRegister
}
