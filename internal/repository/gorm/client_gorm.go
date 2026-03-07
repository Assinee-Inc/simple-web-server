package gorm

import salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"

// ClientGormRepository is a type alias for salesrepogorm.ClientGormRepository for backwards compatibility.
type ClientGormRepository = salesrepogorm.ClientGormRepository

// NewClientGormRepository is a function alias for backwards compatibility.
var NewClientGormRepository = salesrepogorm.NewClientGormRepository

// ContainsNameCpfEmailOrPhoneWith is a function alias for backwards compatibility.
var ContainsNameCpfEmailOrPhoneWith = salesrepogorm.ContainsNameCpfEmailOrPhoneWith
