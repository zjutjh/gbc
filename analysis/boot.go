package analysis

// Init 加载标准库和状态码
func Init() error {
	err := loadStdPackages()
	if err != nil {
		return err
	}
	return loadStatusCodesAndTemplate()
}
