package validate

//发送邮箱验证
type PostEmailValidate struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Type  string `form:"type" json:"type" binding:"required,oneof=register login reset change_email"`
}

// 绑定模型获取验证错误的方法
func (r *PostEmailValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "Incorrect mail format"
	MsgMap["Type.required"] = "type required"
	MsgMap["Type.oneof"] = "type mismatch"
	return ForRangeValidateError(err, MsgMap)
}

//注册验证
type RegisterValidate struct {
	Email         string `form:"email" json:"email" binding:"required,email"`
	Code          string `form:"code" json:"code" binding:"required"`
	Password      string `form:"password" json:"password" binding:"required,min=8,max=20,charNumber"`
	PasswordAgain string `form:"password_again" json:"password_again" binding:"required,eqfield=Password"`
	ReferCode     string `form:"refer_code" json:"refer_code"`
}

func (r *RegisterValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "Incorrect mail format"
	MsgMap["Code.required"] = "code required"
	MsgMap["Password.required"] = "password required"
	MsgMap["Password.min"] = "minimum 8 digits of password"
	MsgMap["Password.max"] = "maximum 20 digits of password"
	MsgMap["Password.charNumber"] = "password must contain upper and lower case letters and numbers"
	MsgMap["PasswordAgain.required"] = "password_again required"
	MsgMap["PasswordAgain.eqfield"] = "two passwords are inconsistent"
	return ForRangeValidateError(err, MsgMap)
}

//验证码登陆校验
type CodeLoginValidate struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Code  string `form:"code" json:"code" binding:"required"`
}

func (r *CodeLoginValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "incorrect mail format"
	MsgMap["Code.required"] = "code required"
	return ForRangeValidateError(err, MsgMap)
}

//密码登陆校验
type PassLoginValidate struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

func (r *PassLoginValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "incorrect mail format"
	MsgMap["Password.required"] = "password required"
	return ForRangeValidateError(err, MsgMap)
}

//GoogleLoginValidate 谷歌登陆校验
type GoogleLoginValidate struct {
	Email       string `form:"email" json:"email" binding:"required,email"`
	AccessToken string `form:"accessToken" json:"accessToken" binding:"required"`
	Name        string `form:"Name" json:"Name" binding:"required"`
	ImageUrl    string `form:"imageUrl" json:"imageUrl"`
}

func (r *GoogleLoginValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "incorrect mail format"
	MsgMap["AccessToken.required"] = "accessToken required"
	MsgMap["Name.required"] = "name required"
	return ForRangeValidateError(err, MsgMap)
}

//重置密码校验
type ResetFirstValidate struct {
	Email string `form:"email" json:"email" binding:"required,email"`
}

func (r *ResetFirstValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "incorrect mail format"
	return ForRangeValidateError(err, MsgMap)
}

//重置密码校验
type ResetPassValidate struct {
	Email         string `form:"email" json:"email" binding:"required,email"`
	Code          string `form:"code" json:"code" binding:"required"`
	Password      string `form:"password" json:"password" binding:"required,min=8,max=20,charNumber"`
	PasswordAgain string `form:"password_again" json:"password_again" binding:"required,eqfield=Password"`
}

func (r *ResetPassValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "Incorrect mail format"
	MsgMap["Code.required"] = "code required"
	MsgMap["Password.required"] = "password required"
	MsgMap["Password.min"] = "minimum 8 digits of password"
	MsgMap["Password.max"] = "maximum 20 digits of password"
	MsgMap["Password.charNumber"] = "password must contain upper and lower case letters and numbers"
	MsgMap["PasswordAgain.required"] = "password_again required"
	MsgMap["PasswordAgain.eqfield"] = "two passwords are inconsistent"
	return ForRangeValidateError(err, MsgMap)
}

//ToResetPassValidate 用户重置密码校验
type ToResetPassValidate struct {
	OriginalPassword string `form:"original_password" json:"original_password" binding:"required"`
	Password         string `form:"password" json:"password" binding:"required,min=8,max=20,charNumber"`
	PasswordAgain    string `form:"password_again" json:"password_again" binding:"required,eqfield=Password"`
}

func (r *ToResetPassValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["OriginalPassword.required"] = "email required"
	MsgMap["Password.required"] = "password required"
	MsgMap["Password.min"] = "minimum 8 digits of password"
	MsgMap["Password.max"] = "maximum 20 digits of password"
	MsgMap["Password.charNumber"] = "password must contain upper and lower case letters and numbers"
	MsgMap["PasswordAgain.required"] = "password_again required"
	MsgMap["PasswordAgain.eqfield"] = "two passwords are inconsistent"
	return ForRangeValidateError(err, MsgMap)
}

//ChangeEmailValidate 用户更换邮箱校验
type ChangeEmailValidate struct {
	Password string `form:"password" json:"password" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required,email"`
	Code     string `form:"code" json:"code" binding:"required"`
}

func (r *ChangeEmailValidate) GetError(err error) string {
	MsgMap := map[string]string{}
	MsgMap["Password.required"] = "password required"
	MsgMap["Email.required"] = "email required"
	MsgMap["Email.email"] = "incorrect mail format"
	MsgMap["Code.required"] = "code required"
	return ForRangeValidateError(err, MsgMap)
}
