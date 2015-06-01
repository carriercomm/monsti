// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann <cneumann@datenkarussell.de>
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"path"
	"github.com/chrneumann/htmlwidgets"
	gomail "gopkg.in/gomail.v1"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/util/i18n"
	"pkg.monsti.org/monsti/api/util/template"
)
import "pkg.monsti.org/monsti/api/service"

var availableLocales = []string{"en", "de"}

func initNodeTypes(settings *settings, session *service.Session, logger *log.Logger) error {
	G := func(in string) string { return in }
	pathType := service.NodeType{
		Id:   "core.Path",
		Hide: true,
		Name: i18n.GenLanguageMap(G("Path"), availableLocales),
	}
	if err := session.Monsti().RegisterNodeType(&pathType); err != nil {
		return fmt.Errorf("Could not register path node type: %v", err)
	}

	documentType := service.NodeType{
		Id:        "core.Document",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Document"), availableLocales),
		Fields: []*service.FieldConfig{
			{
				Id:       "core.Title",
				Required: true,
				Name:     i18n.GenLanguageMap(G("Title"), availableLocales),
				Type:     new(service.TextFieldType),
			},
			{
				Id:   "core.Description",
				Name: i18n.GenLanguageMap(G("Description"), availableLocales),
				Type: new(service.TextFieldType),
			},
			{
				Id:   "core.Thumbnail",
				Name: i18n.GenLanguageMap(G("Thumbnail"), availableLocales),
				Type: new(service.RefFieldType),
			},
			{
				Id:       "core.Body",
				Required: true,
				Name:     i18n.GenLanguageMap(G("Body"), availableLocales),
				Type:     new(service.HTMLFieldType),
			},
		},
	}
	if err := session.Monsti().RegisterNodeType(&documentType); err != nil {
		return fmt.Errorf("Could not register document node type: %v", err)
	}

	fileType := service.NodeType{
		Id:        "core.File",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("File"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{
				Id:       "core.File",
				Required: true,
				Name:     i18n.GenLanguageMap(G("File"), availableLocales),
				Type:     new(service.FileFieldType),
			},
		},
	}
	if err := session.Monsti().RegisterNodeType(&fileType); err != nil {
		return fmt.Errorf("Could not register file node type: %v", err)
	}

	imageType := service.NodeType{
		Id:        "core.Image",
		Hide:      true,
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Image"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{Id: "core.File"},
		},
	}
	if err := session.Monsti().RegisterNodeType(&imageType); err != nil {
		return fmt.Errorf("Could not register image node type: %v", err)
	}

	contactFormType := service.NodeType{
		Id:        "core.ContactForm",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Contact form"), availableLocales),
		Fields:    service.CoreFields,
	}
	if err := session.Monsti().RegisterNodeType(&contactFormType); err != nil {
		return fmt.Errorf("Could not register contactform node type: %v", err)
	}
	return nil
}

type contactFormData struct {
	Name, Email, Subject, Message string
}

func renderContactForm(c *reqContext, context template.Context,
	formValues url.Values, h *nodeHandler) error {
	G, _, _, _ := gettext.DefaultLocales.Use("",
		c.SiteSettings.Fields["core.Locale"].Value().(string))
	m := c.Serv.Monsti()
	data := contactFormData{}
	form := htmlwidgets.NewForm(&data)
	form.AddWidget(&htmlwidgets.TextWidget{MinLength: 1,
		ValidationError: G("Required.")}, "Name", G("Name"), "")
	form.AddWidget(&htmlwidgets.TextWidget{MinLength: 1,
		ValidationError: G("Required.")}, "Email", G("Email"), "")
	form.AddWidget(&htmlwidgets.TextWidget{MinLength: 1,
		ValidationError: G("Required.")}, "Subject", G("Subject"), "")
	form.AddWidget(&htmlwidgets.TextAreaWidget{MinLength: 1,
		ValidationError: G("Required.")}, "Message", G("Message"), "")

	switch c.Req.Method {
	case "GET":
		if _, submitted := formValues["submitted"]; submitted {
			context["Submitted"] = 1
		}
	case "POST":
		if form.Fill(formValues) {
			mail := gomail.NewMessage()
			mail.SetAddressHeader("From",
				c.SiteSettings.StringValue("core.EmailAddress"),
				c.SiteSettings.StringValue("core.EmailName"))
			mail.SetAddressHeader("To",
				c.SiteSettings.StringValue("core.OwnerEmail"),
				c.SiteSettings.StringValue("core.OwnerName"))
			mail.SetAddressHeader("Reply-To", data.Email, data.Name)
			mail.SetHeader("Subject", data.Subject)
			body := fmt.Sprintf("%v\n%v\n\n%v",
				fmt.Sprintf(G("Received from contact form at %v"),
					c.SiteSettings.StringValue("core.Title")),
				fmt.Sprintf(G("Name: %v | Email: %v"), data.Name, data.Email),
				data.Message)
			mail.SetBody("text/plain", body)
			mailer := gomail.NewCustomMailer("", nil, gomail.SetSendMail(
				m.SendMailFunc()))
			err := mailer.Send(mail)
			if err != nil {
				return fmt.Errorf("Could not send mail: %v", err)
			}
			http.Redirect(c.Res, c.Req, path.Dir(c.Node.Path)+"/?submitted", http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	context["Form"] = form.RenderData()
	return nil
}
