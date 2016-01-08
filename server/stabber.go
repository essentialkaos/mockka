package server

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"net/http"
	"strings"

	"github.com/icrowley/fake"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Stabber struct
type Stabber struct {
	request *http.Request
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Last used fake language
var stabberFakeLang = "en"

// ////////////////////////////////////////////////////////////////////////////////// //

// Query return query value from request
func (s *Stabber) Query(name string) string {
	if s.request == nil {
		return ""
	}

	query := s.request.URL.Query()

	return strings.Join(query[name], " ")
}

// QueryIs if query have given value
func (s *Stabber) QueryIs(name, value string) bool {
	return s.Query(name) == value
}

// Header return header value from request
func (s *Stabber) Header(name string) string {
	if s.request == nil {
		return ""
	}

	headers := s.request.Header

	return strings.Join(headers[name], " ")
}

// HeaderIs if header have given value
func (s *Stabber) HeaderIs(name, value string) bool {
	return s.Header(name) == value
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Brand generates brand name
func (s *Stabber) Brand(lang string) string {
	setLang(lang)
	return fake.Brand()
}

// Character generates random character in the given language
func (s *Stabber) Character(lang string) string {
	setLang(lang)
	return fake.Character()
}

// Characters generates from 1 to 5 characters in the given language
func (s *Stabber) Characters(lang string) string {
	setLang(lang)
	return fake.Characters()
}

// CharactersN generates n random characters in the given language
func (s *Stabber) CharactersN(lang string, n int) string {
	setLang(lang)
	return fake.CharactersN(n)
}

// City generates random city
func (s *Stabber) City(lang string) string {
	setLang(lang)
	return fake.City()
}

// Color generates color name
func (s *Stabber) Color(lang string) string {
	setLang(lang)
	return fake.Color()
}

// Company generates company name
func (s *Stabber) Company(lang string) string {
	setLang(lang)
	return fake.Company()
}

// Continent generates random continent
func (s *Stabber) Continent(lang string) string {
	setLang(lang)
	return fake.Continent()
}

// Country generates random country
func (s *Stabber) Country(lang string) string {
	setLang(lang)
	return fake.Country()
}

// CreditCardNum generated credit card number according to the card number rules
func (s *Stabber) CreditCardNum(vendor string) string {
	return fake.CreditCardNum(vendor)
}

// CreditCardType returns one of the following credit values:
// VISA, MasterCard, American Express and Discover
func (s *Stabber) CreditCardType() string {
	return fake.CreditCardType()
}

// Currency generates currency name
func (s *Stabber) Currency(lang string) string {
	setLang(lang)
	return fake.Currency()
}

// CurrencyCode generates currency code
func (s *Stabber) CurrencyCode() string {
	return fake.CurrencyCode()
}

// Day generates day of the month
func (s *Stabber) Day() int {
	return fake.Day()
}

// Digits returns from 1 to 5 digits as a string
func (s *Stabber) Digits(lang string) string {
	setLang(lang)
	return fake.Digits()
}

// DigitsN returns n digits as a string
func (s *Stabber) DigitsN(lang string, n int) string {
	setLang(lang)
	return fake.DigitsN(n)
}

// DomainName generates random domain name
func (s *Stabber) DomainName() string {
	return fake.DomainName()
}

// DomainZone generates random domain zone
func (s *Stabber) DomainZone() string {
	return fake.DomainZone()
}

// EmailAddress generates email address
func (s *Stabber) EmailAddress() string {
	return fake.EmailAddress()
}

// EmailBody generates random email body
func (s *Stabber) EmailBody(lang string) string {
	setLang(lang)
	return fake.EmailBody()
}

// EmailSubject generates random email subject
func (s *Stabber) EmailSubject(lang string) string {
	setLang(lang)
	return fake.EmailSubject()
}

// FemaleFirstName generates female first name
func (s *Stabber) FemaleFirstName(lang string) string {
	setLang(lang)
	return fake.FemaleFirstName()
}

// FemaleFullName generates female full name it can occasionally
// include prefix or suffix
func (s *Stabber) FemaleFullName(lang string) string {
	setLang(lang)
	return fake.FemaleFullName()
}

// FemaleFullNameWithPrefix generates prefixed female full name
// if prefixes for the given language are available
func (s *Stabber) FemaleFullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.FemaleFullNameWithPrefix()
}

// FemaleFullNameWithSuffix generates suffixed female full name
// if suffixes for the given language are available
func (s *Stabber) FemaleFullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.FemaleFullNameWithSuffix()
}

// FemaleLastName generates female last name
func (s *Stabber) FemaleLastName(lang string) string {
	setLang(lang)
	return fake.FemaleLastName()
}

// FemalePatronymic generates female patronymic
func (s *Stabber) FemalePatronymic(lang string) string {
	setLang(lang)
	return fake.FemalePatronymic()
}

// FirstName generates first name
func (s *Stabber) FirstName(lang string) string {
	setLang(lang)
	return fake.FirstName()
}

// FullName generates full name it can occasionally include prefix
// or suffix
func (s *Stabber) FullName(lang string) string {
	setLang(lang)
	return fake.FullName()
}

// FullNameWithPrefix generates prefixed full name if prefixes for
// the given language are available
func (s *Stabber) FullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.FullNameWithPrefix()
}

// FullNameWithSuffix generates suffixed full name if suffixes for
// the given language are available
func (s *Stabber) FullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.FullNameWithSuffix()
}

// Gender generates random gender
func (s *Stabber) Gender(lang string) string {
	setLang(lang)
	return fake.Gender()
}

// GenderAbbrev returns first downcased letter of the random gender
func (s *Stabber) GenderAbbrev(lang string) string {
	setLang(lang)
	return fake.GenderAbbrev()
}

// HexColor generates hex color name
func (s *Stabber) HexColor() string {
	return fake.HexColor()
}

// HexColorShort generates short hex color name
func (s *Stabber) HexColorShort() string {
	return fake.HexColorShort()
}

// IPv4 generates IPv4 address
func (s *Stabber) IPv4() string {
	return fake.IPv4()
}

// Industry generates industry name
func (s *Stabber) Industry(lang string) string {
	setLang(lang)
	return fake.Industry()
}

// JobTitle generates job title
func (s *Stabber) JobTitle(lang string) string {
	setLang(lang)
	return fake.JobTitle()
}

// Language generates random human language
func (s *Stabber) Language(lang string) string {
	setLang(lang)
	return fake.Language()
}

// LastName generates last name
func (s *Stabber) LastName(lang string) string {
	setLang(lang)
	return fake.LastName()
}

// LatitudeDegress generates latitude degrees (from -180 to 180)
func (s *Stabber) LatitudeDegress() int {
	return fake.LatitudeDegress()
}

// LatitudeDirection generates latitude direction (N(orth) o S(outh))
func (s *Stabber) LatitudeDirection(lang string) string {
	setLang(lang)
	return fake.LatitudeDirection()
}

// LatitudeMinutes generates latitude minutes (from 0 to 60)
func (s *Stabber) LatitudeMinutes() int {
	return fake.LatitudeMinutes()
}

// LatitudeSeconds generates latitude seconds (from 0 to 60)
func (s *Stabber) LatitudeSeconds() int {
	return fake.LatitudeSeconds()
}

// Latitute generates latitude
func (s *Stabber) Latitute() float32 {
	return fake.Latitute()
}

// Longitude generates longitude
func (s *Stabber) Longitude() float32 {
	return fake.Longitude()
}

// LongitudeDegrees generates longitude degrees (from -180 to 180)
func (s *Stabber) LongitudeDegrees() int {
	return fake.LongitudeDegrees()
}

// LongitudeDirection generates (W(est) or E(ast))
func (s *Stabber) LongitudeDirection(lang string) string {
	setLang(lang)
	return fake.LongitudeDirection()
}

// LongitudeMinutes generates (from 0 to 60)
func (s *Stabber) LongitudeMinutes() int {
	return fake.LongitudeMinutes()
}

// LongitudeSeconds generates (from 0 to 60)
func (s *Stabber) LongitudeSeconds() int {
	return fake.LongitudeSeconds()
}

// MaleFirstName generates male first name
func (s *Stabber) MaleFirstName(lang string) string {
	setLang(lang)
	return fake.MaleFirstName()
}

// MaleFullName generates male full name it can occasionally include prefix
// or suffix
func (s *Stabber) MaleFullName(lang string) string {
	setLang(lang)
	return fake.MaleFullName()
}

// MaleFullNameWithPrefix generates prefixed male full name if prefixes for
// the given language are available
func (s *Stabber) MaleFullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.MaleFullNameWithPrefix()
}

// MaleFullNameWithSuffix generates suffixed male full name if suffixes for
// the given language are available
func (s *Stabber) MaleFullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.MaleFullNameWithSuffix()
}

// MaleLastName generates male last name
func (s *Stabber) MaleLastName(lang string) string {
	setLang(lang)
	return fake.MaleLastName()
}

// MalePatronymic generates male patronymic
func (s *Stabber) MalePatronymic(lang string) string {
	setLang(lang)
	return fake.MalePatronymic()
}

// Model generates model name that consists of letters and digits, optionally
// with a hyphen between them
func (s *Stabber) Model(lang string) string {
	setLang(lang)
	return fake.Model()
}

// Month generates month name
func (s *Stabber) Month(lang string) string {
	setLang(lang)
	return fake.Month()
}

// MonthNum generates month number (from 1 to 12)
func (s *Stabber) MonthNum() int {
	return fake.MonthNum()
}

// MonthShort generates abbreviated month name
func (s *Stabber) MonthShort(lang string) string {
	setLang(lang)
	return fake.MonthShort()
}

// Paragraph generates paragraph
func (s *Stabber) Paragraph(lang string) string {
	setLang(lang)
	return fake.Paragraph()
}

// Paragraphs generates from 1 to 5 paragraphs
func (s *Stabber) Paragraphs(lang string) string {
	setLang(lang)
	return fake.Paragraphs()
}

// ParagraphsN generates n paragraphs
func (s *Stabber) ParagraphsN(lang string, n int) string {
	setLang(lang)
	return fake.ParagraphsN(n)
}

// Password generates password with the length from atLeast to atMOst charachers,
// allow* parameters specify whether corresponding symbols can be used
func (s *Stabber) Password(atLeast, atMost int, allowUpper, allowNumeric, allowSpecial bool) string {
	return fake.Password(atLeast, atMost, allowUpper, allowNumeric, allowSpecial)
}

// Patronymic generates patronymic
func (s *Stabber) Patronymic(lang string) string {
	setLang(lang)
	return fake.Patronymic()
}

// Phone generates random phone number using one of the formats format
// specified in phone_format file
func (s *Stabber) Phone(lang string) string {
	setLang(lang)
	return fake.Phone()
}

// Product generates product title as brand + product name
func (s *Stabber) Product(lang string) string {
	setLang(lang)
	return fake.Product()
}

// ProductName generates product name
func (s *Stabber) ProductName(lang string) string {
	setLang(lang)
	return fake.ProductName()
}

// Sentence generates random sentence
func (s *Stabber) Sentence(lang string) string {
	setLang(lang)
	return fake.Sentence()
}

// Sentences generates from 1 to 5 random sentences
func (s *Stabber) Sentences(lang string) string {
	setLang(lang)
	return fake.Sentences()
}

// SentencesN generates n random sentences
func (s *Stabber) SentencesN(lang string, n int) string {
	setLang(lang)
	return fake.SentencesN(n)
}

// SimplePassword is a wrapper around Password, it generates password with the length
// from 6 to 12 symbols, with upper characters and numeric symbols allowed
func (s *Stabber) SimplePassword() string {
	return fake.SimplePassword()
}

// State generates random state
func (s *Stabber) State(lang string) string {
	setLang(lang)
	return fake.State()
}

// StateAbbrev generates random state abbreviation
func (s *Stabber) StateAbbrev(lang string) string {
	setLang(lang)
	return fake.StateAbbrev()
}

// Street generates random street name
func (s *Stabber) Street(lang string) string {
	setLang(lang)
	return fake.Street()
}

// StreetAddress generates random street name along with building number
func (s *Stabber) StreetAddress(lang string) string {
	setLang(lang)
	return fake.StreetAddress()
}

// Title generates from 2 to 5 titleized words
func (s *Stabber) Title(lang string) string {
	setLang(lang)
	return fake.Title()
}

// TopLevelDomain generates random top level domain
func (s *Stabber) TopLevelDomain() string {
	return fake.TopLevelDomain()
}

// UserName generates user name in one of the following forms first name + last
// name, letter + last names or concatenation of from 1 to 3 lowercased words
func (s *Stabber) UserName(lang string) string {
	fake.SetLang("en")

	username := fake.UserName()

	fake.SetLang(stabberFakeLang)

	return username
}

// WeekDay generates name ot the week day
func (s *Stabber) WeekDay(lang string) string {
	setLang(lang)
	return fake.WeekDay()
}

// WeekDayShort generates abbreviated name of the week day
func (s *Stabber) WeekDayShort(lang string) string {
	setLang(lang)
	return fake.WeekDayShort()
}

// WeekdayNum generates number of the day of the week
func (s *Stabber) WeekdayNum(lang string) int {
	return fake.WeekdayNum()
}

// Word generates random word
func (s *Stabber) Word(lang string) string {
	setLang(lang)
	return fake.Word()
}

// Words generates from 1 to 5 random words
func (s *Stabber) Words(lang string) string {
	setLang(lang)
	return fake.Words()
}

// WordsN generates n random words
func (s *Stabber) WordsN(lang string, n int) string {
	setLang(lang)
	return fake.WordsN(n)
}

// Year generates year using the given boundaries
func (s *Stabber) Year(from, to int) int {
	return fake.Year(from, to)
}

// Zip generates random zip code using one of the formats specifies in zip_format file
func (s *Stabber) Zip() string {
	return fake.Zip()
}

// ////////////////////////////////////////////////////////////////////////////////// //

func setLang(lang string) {
	if lang != "" && lang != stabberFakeLang {
		fake.SetLang(lang)
	}
}
