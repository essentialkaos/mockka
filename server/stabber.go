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

// Last used fake language
var stabberFakeLang string = "en"

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *Stabber) Query(name string) string {
	if s.request == nil {
		return ""
	}

	query := s.request.URL.Query()

	return strings.Join(query[name], " ")
}

func (s *Stabber) QueryIs(name, value string) bool {
	return s.Query(name) == value
}

func (s *Stabber) Header(name string) string {
	if s.request == nil {
		return ""
	}

	headers := s.request.Header

	return strings.Join(headers[name], " ")
}

func (s *Stabber) HeaderIs(name, value string) bool {
	return s.Header(name) == value
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *Stabber) Brand(lang string) string {
	setLang(lang)
	return fake.Brand()
}

func (s *Stabber) Character(lang string) string {
	setLang(lang)
	return fake.Character()
}

func (s *Stabber) Characters(lang string) string {
	setLang(lang)
	return fake.Characters()
}

func (s *Stabber) CharactersN(lang string, n int) string {
	setLang(lang)
	return fake.CharactersN(n)
}

func (s *Stabber) City(lang string) string {
	setLang(lang)
	return fake.City()
}

func (s *Stabber) Color(lang string) string {
	setLang(lang)
	return fake.Color()
}

func (s *Stabber) Company(lang string) string {
	setLang(lang)
	return fake.Company()
}

func (s *Stabber) Continent(lang string) string {
	setLang(lang)
	return fake.Continent()
}

func (s *Stabber) Country(lang string) string {
	setLang(lang)
	return fake.Country()
}

func (s *Stabber) CreditCardNum(vendor string) string {
	return fake.CreditCardNum(vendor)
}

func (s *Stabber) CreditCardType() string {
	return fake.CreditCardType()
}

func (s *Stabber) Currency(lang string) string {
	setLang(lang)
	return fake.Currency()
}

func (s *Stabber) CurrencyCode() string {
	return fake.CurrencyCode()
}

func (s *Stabber) Day(lang string) int {
	setLang(lang)
	return fake.Day()
}

func (s *Stabber) Digits(lang string) string {
	setLang(lang)
	return fake.Digits()
}

func (s *Stabber) DigitsN(lang string, n int) string {
	setLang(lang)
	return fake.DigitsN(n)
}

func (s *Stabber) DomainName() string {
	return fake.DomainName()
}

func (s *Stabber) DomainZone() string {
	return fake.DomainZone()
}

func (s *Stabber) EmailAddress() string {
	return fake.EmailAddress()
}

func (s *Stabber) EmailBody(lang string) string {
	setLang(lang)
	return fake.EmailBody()
}

func (s *Stabber) EmailSubject(lang string) string {
	setLang(lang)
	return fake.EmailSubject()
}

func (s *Stabber) FemaleFirstName(lang string) string {
	setLang(lang)
	return fake.FemaleFirstName()
}

func (s *Stabber) FemaleFullName(lang string) string {
	setLang(lang)
	return fake.FemaleFullName()
}

func (s *Stabber) FemaleFullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.FemaleFullNameWithPrefix()
}

func (s *Stabber) FemaleFullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.FemaleFullNameWithSuffix()
}

func (s *Stabber) FemaleLastName(lang string) string {
	setLang(lang)
	return fake.FemaleLastName()
}

func (s *Stabber) FemalePatronymic(lang string) string {
	setLang(lang)
	return fake.FemalePatronymic()
}

func (s *Stabber) FirstName(lang string) string {
	setLang(lang)
	return fake.FirstName()
}

func (s *Stabber) FullName(lang string) string {
	setLang(lang)
	return fake.FullName()
}

func (s *Stabber) FullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.FullNameWithPrefix()
}

func (s *Stabber) FullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.FullNameWithSuffix()
}

func (s *Stabber) Gender(lang string) string {
	setLang(lang)
	return fake.Gender()
}

func (s *Stabber) GenderAbbrev(lang string) string {
	setLang(lang)
	return fake.GenderAbbrev()
}

func (s *Stabber) HexColor() string {
	return fake.HexColor()
}

func (s *Stabber) HexColorShort() string {
	return fake.HexColorShort()
}

func (s *Stabber) IPv4() string {
	return fake.IPv4()
}

func (s *Stabber) Industry(lang string) string {
	setLang(lang)
	return fake.Industry()
}

func (s *Stabber) JobTitle(lang string) string {
	setLang(lang)
	return fake.JobTitle()
}

func (s *Stabber) Language(lang string) string {
	setLang(lang)
	return fake.Language()
}

func (s *Stabber) LastName(lang string) string {
	setLang(lang)
	return fake.LastName()
}

func (s *Stabber) LatitudeDegress() int {
	return fake.LatitudeDegress()
}

func (s *Stabber) LatitudeDirection(lang string) string {
	setLang(lang)
	return fake.LatitudeDirection()
}

func (s *Stabber) LatitudeMinutes() int {
	return fake.LatitudeMinutes()
}

func (s *Stabber) LatitudeSeconds() int {
	return fake.LatitudeSeconds()
}

func (s *Stabber) Latitute() float32 {
	return fake.Latitute()
}

func (s *Stabber) Longitude() float32 {
	return fake.Longitude()
}

func (s *Stabber) LongitudeDegrees() int {
	return fake.LongitudeDegrees()
}

func (s *Stabber) LongitudeDirection(lang string) string {
	setLang(lang)
	return fake.LongitudeDirection()
}

func (s *Stabber) LongitudeMinutes() int {
	return fake.LongitudeMinutes()
}

func (s *Stabber) LongitudeSeconds() int {
	return fake.LongitudeSeconds()
}

func (s *Stabber) MaleFirstName(lang string) string {
	setLang(lang)
	return fake.MaleFirstName()
}

func (s *Stabber) MaleFullName(lang string) string {
	setLang(lang)
	return fake.MaleFullName()
}

func (s *Stabber) MaleFullNameWithPrefix(lang string) string {
	setLang(lang)
	return fake.MaleFullNameWithPrefix()
}

func (s *Stabber) MaleFullNameWithSuffix(lang string) string {
	setLang(lang)
	return fake.MaleFullNameWithSuffix()
}

func (s *Stabber) MaleLastName(lang string) string {
	setLang(lang)
	return fake.MaleLastName()
}

func (s *Stabber) MalePatronymic(lang string) string {
	setLang(lang)
	return fake.MalePatronymic()
}

func (s *Stabber) Model(lang string) string {
	setLang(lang)
	return fake.Model()
}

func (s *Stabber) Month(lang string) string {
	setLang(lang)
	return fake.Month()
}

func (s *Stabber) MonthNum() int {
	return fake.MonthNum()
}

func (s *Stabber) MonthShort(lang string) string {
	setLang(lang)
	return fake.MonthShort()
}

func (s *Stabber) Paragraph(lang string) string {
	setLang(lang)
	return fake.Paragraph()
}

func (s *Stabber) Paragraphs(lang string) string {
	setLang(lang)
	return fake.Paragraphs()
}

func (s *Stabber) ParagraphsN(lang string, n int) string {
	setLang(lang)
	return fake.ParagraphsN(n)
}

func (s *Stabber) Password(atLeast, atMost int, allowUpper, allowNumeric, allowSpecial bool) string {
	return fake.Password(atLeast, atMost, allowUpper, allowNumeric, allowSpecial)
}

func (s *Stabber) Patronymic(lang string) string {
	setLang(lang)
	return fake.Patronymic()
}

func (s *Stabber) Phone(lang string) string {
	setLang(lang)
	return fake.Phone()
}

func (s *Stabber) Product(lang string) string {
	setLang(lang)
	return fake.Product()
}

func (s *Stabber) ProductName(lang string) string {
	setLang(lang)
	return fake.ProductName()
}

func (s *Stabber) Sentence(lang string) string {
	setLang(lang)
	return fake.Sentence()
}

func (s *Stabber) Sentences(lang string) string {
	setLang(lang)
	return fake.Sentences()
}

func (s *Stabber) SentencesN(lang string, n int) string {
	setLang(lang)
	return fake.SentencesN(n)
}

func (s *Stabber) SimplePassword() string {
	return fake.SimplePassword()
}

func (s *Stabber) State(lang string) string {
	setLang(lang)
	return fake.State()
}

func (s *Stabber) StateAbbrev(lang string) string {
	setLang(lang)
	return fake.StateAbbrev()
}

func (s *Stabber) Street(lang string) string {
	setLang(lang)
	return fake.Street()
}

func (s *Stabber) StreetAddress(lang string) string {
	setLang(lang)
	return fake.StreetAddress()
}

func (s *Stabber) Title(lang string) string {
	setLang(lang)
	return fake.Title()
}

func (s *Stabber) TopLevelDomain() string {
	return fake.TopLevelDomain()
}

func (s *Stabber) UserName(lang string) string {
	fake.SetLang("en")

	username := fake.UserName()

	fake.SetLang(stabberFakeLang)

	return username
}

func (s *Stabber) WeekDay(lang string) string {
	setLang(lang)
	return fake.WeekDay()
}

func (s *Stabber) WeekDayShort(lang string) string {
	setLang(lang)
	return fake.WeekDayShort()
}

func (s *Stabber) WeekdayNum(lang string) int {
	setLang(lang)
	return fake.WeekdayNum()
}

func (s *Stabber) Word(lang string) string {
	setLang(lang)
	return fake.Word()
}

func (s *Stabber) Words(lang string) string {
	setLang(lang)
	return fake.Words()
}

func (s *Stabber) WordsN(lang string, n int) string {
	setLang(lang)
	return fake.WordsN(n)
}

func (s *Stabber) Year(from, to int) int {
	return fake.Year(from, to)
}

func (s *Stabber) Zip() string {
	return fake.Zip()
}

// ////////////////////////////////////////////////////////////////////////////////// //

func setLang(lang string) {
	if lang != "" && lang != stabberFakeLang {
		fake.SetLang(lang)
	}
}
