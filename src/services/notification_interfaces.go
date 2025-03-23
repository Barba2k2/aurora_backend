package services

type EmailServiceInterface interface {
	// SendPasswordResetEmail envia um email de recuperação de senha
	// Parâmetros:
	// - email: endereço de email do destinatário
	// - name: nome do destinatário para personalização
	// - token: token de recuperação de senha
	SendPasswordResetEmail(email, name, token string) error

	// SendGenericEmail envia um email genérico
	// Parâmetros:
	// - email: endereço de email do destinatário
	// - subject: assunto do email
	// - body: corpo do email (HTML ou texto)
	SendGenericEmail(email, subject, body string) error
	
	// SendAppointmentConfirmation envia um email de confirmação de agendamento
	// Parâmetros:
	// - email: endereço de email do destinatário
	// - name: nome do destinatário
	// - appointmentID: ID do agendamento
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	// - professionalName: nome do profissional
	SendAppointmentConfirmation(email, name, appointmentID, serviceName, dateTime, professionalName string) error
	
	// SendAppointmentReminder envia um lembrete de agendamento
	// Parâmetros:
	// - email: endereço de email do destinatário
	// - name: nome do destinatário
	// - appointmentID: ID do agendamento
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	// - professionalName: nome do profissional
	SendAppointmentReminder(email, name, appointmentID, serviceName, dateTime, professionalName string) error
	
	// SendAppointmentCancellation envia uma notificação de cancelamento de agendamento
	// Parâmetros:
	// - email: endereço de email do destinatário
	// - name: nome do destinatário
	// - serviceName: nome do serviço que foi cancelado
	// - dateTime: data e hora que estava agendada
	// - cancellationReason: motivo do cancelamento (opcional)
	SendAppointmentCancellation(email, name, serviceName, dateTime, cancellationReason string) error
}

// SMSServiceInterface define a interface para o serviço de envio de SMS
type SMSServiceInterface interface {
	// SendPasswordResetSMS envia um SMS com código de recuperação de senha
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - code: código de recuperação de senha
	SendPasswordResetSMS(phone, code string) error
	
	// SendGenericSMS envia um SMS genérico
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - message: mensagem a ser enviada
	SendGenericSMS(phone, message string) error
	
	// SendAppointmentConfirmationSMS envia um SMS de confirmação de agendamento
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - appointmentID: ID do agendamento
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	SendAppointmentConfirmationSMS(phone, appointmentID, serviceName, dateTime string) error
	
	// SendAppointmentReminderSMS envia um lembrete de agendamento por SMS
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	SendAppointmentReminderSMS(phone, serviceName, dateTime string) error
	
	// SendAppointmentCancellationSMS envia uma notificação de cancelamento por SMS
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - serviceName: nome do serviço que foi cancelado
	// - dateTime: data e hora que estava agendada
	SendAppointmentCancellationSMS(phone, serviceName, dateTime string) error
}

// WhatsAppServiceInterface define a interface para o serviço de envio de mensagens WhatsApp
type WhatsAppServiceInterface interface {
	// SendPasswordResetWhatsApp envia uma mensagem de WhatsApp com código de recuperação de senha
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - name: nome do destinatário para personalização
	// - code: código de recuperação de senha
	SendPasswordResetWhatsApp(phone, name, code string) error
	
	// SendGenericWhatsApp envia uma mensagem de WhatsApp genérica
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - message: mensagem a ser enviada
	SendGenericWhatsApp(phone, message string) error
	
	// SendAppointmentConfirmationWhatsApp envia confirmação de agendamento por WhatsApp
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - name: nome do destinatário
	// - appointmentID: ID do agendamento
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	// - professionalName: nome do profissional
	SendAppointmentConfirmationWhatsApp(phone, name, appointmentID, serviceName, dateTime, professionalName string) error
	
	// SendAppointmentReminderWhatsApp envia lembrete de agendamento por WhatsApp
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - name: nome do destinatário
	// - serviceName: nome do serviço agendado
	// - dateTime: data e hora do agendamento
	// - professionalName: nome do profissional
	SendAppointmentReminderWhatsApp(phone, name, serviceName, dateTime, professionalName string) error
	
	// SendAppointmentCancellationWhatsApp envia notificação de cancelamento por WhatsApp
	// Parâmetros:
	// - phone: número de telefone do destinatário
	// - name: nome do destinatário
	// - serviceName: nome do serviço que foi cancelado
	// - dateTime: data e hora que estava agendada
	// - cancellationReason: motivo do cancelamento (opcional)
	SendAppointmentCancellationWhatsApp(phone, name, serviceName, dateTime, cancellationReason string) error
}

// PushNotificationServiceInterface define a interface para o serviço de notificações push
type PushNotificationServiceInterface interface {
	// SendPushNotification envia uma notificação push para o dispositivo do usuário
	// Parâmetros:
	// - subscription: assinatura de push notification do usuário
	// - title: título da notificação
	// - body: corpo da notificação
	// - data: dados adicionais para a notificação (opcional)
	SendPushNotification(subscription string, title, body string, data map[string]interface{}) error
	
	// SavePushSubscription salva uma nova assinatura de push notification
	// Parâmetros:
	// - userID: ID do usuário
	// - subscription: assinatura de push notification
	SavePushSubscription(userID string, subscription string) error
	
	// RemovePushSubscription remove uma assinatura de push notification
	// Parâmetros:
	// - userID: ID do usuário
	// - subscription: assinatura de push notification
	RemovePushSubscription(userID string, subscription string) error
}

// NotificationProcessorInterface define a interface para o processador de notificações
// que coordena o envio de notificações por múltiplos canais com fallback automático
type NotificationProcessorInterface interface {
	// SendAppointmentNotification envia notificação relacionada a agendamento por todos os canais disponíveis
	// Parâmetros:
	// - userID: ID do usuário
	// - notificationType: tipo de notificação (confirmation, reminder, cancellation)
	// - appointmentData: dados do agendamento
	// - preferredChannels: canais preferidos para envio, em ordem de prioridade
	SendAppointmentNotification(userID string, notificationType string, appointmentData map[string]string, preferredChannels []string) error
	
	// SendPasswordResetNotification envia notificação de recuperação de senha por canal específico
	// Parâmetros:
	// - userID: ID do usuário
	// - channel: canal de envio (email, sms, whatsapp)
	// - tokenData: dados do token de recuperação
	SendPasswordResetNotification(userID string, channel string, tokenData map[string]string) error
}