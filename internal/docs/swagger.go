package docs

import "encoding/json"

func SpecJSON() ([]byte, error) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Tramplin API",
			"version":     "1.0.0",
			"description": "REST API интерактивной карьерной платформы на Go + Fiber.",
		},
		"servers": []map[string]any{{"url": "/"}},
		"tags": []map[string]any{
			{"name": "health"},
			{"name": "auth"},
			{"name": "public"},
			{"name": "student"},
			{"name": "employer"},
			{"name": "curator"},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"UserHeader": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "X-User-ID",
				},
			},
			"schemas": map[string]any{
				"Envelope": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{"type": "string"},
						"data":   map[string]any{"type": "object", "additionalProperties": true},
						"error":  map[string]any{"type": "string"},
					},
				},
				"RegisterInput": objectSchema(
					prop("email", "string"),
					prop("password", "string"),
					prop("display_name", "string"),
					prop("role", "string"),
					prop("company_name", "string"),
				),
				"LoginInput": objectSchema(
					prop("email", "string"),
					prop("password", "string"),
				),
				"StudentProfileInput": objectSchema(
					prop("last_name", "string"),
					prop("first_name", "string"),
					prop("middle_name", "string"),
					prop("university_name", "string"),
					prop("faculty", "string"),
					prop("specialization", "string"),
					prop("study_year", "integer"),
					prop("graduation_year", "integer"),
					prop("about", "string"),
					prop("profile_visibility", "string"),
				),
				"ResumeInput": objectSchema(
					prop("title", "string"),
					prop("summary", "string"),
					prop("experience_text", "string"),
					prop("education_text", "string"),
				),
				"PortfolioProjectInput": objectSchema(
					prop("title", "string"),
					prop("description", "string"),
					prop("project_url", "string"),
					prop("repository_url", "string"),
					prop("demo_url", "string"),
				),
				"ApplicationInput": objectSchema(
					prop("resume_id", "string"),
					prop("cover_letter", "string"),
				),
				"ContactRequestInput": objectSchema(
					prop("receiver_user_id", "string"),
					prop("message", "string"),
					prop("status", "string"),
				),
				"RecommendationInput": objectSchema(
					prop("to_user_id", "string"),
					prop("opportunity_id", "string"),
					prop("message", "string"),
				),
				"CompanyInput": objectSchema(
					prop("legal_name", "string"),
					prop("brand_name", "string"),
					prop("description", "string"),
					prop("industry", "string"),
					prop("website_url", "string"),
					prop("email_domain", "string"),
					prop("inn", "string"),
					prop("ogrn", "string"),
				),
				"CompanyLinkInput": objectSchema(
					prop("link_type", "string"),
					prop("url", "string"),
				),
				"VerificationInput": objectSchema(
					prop("verification_method", "string"),
					prop("corporate_email", "string"),
					prop("inn_submitted", "string"),
					prop("documents_comment", "string"),
				),
				"OpportunityInput": objectSchema(
					prop("title", "string"),
					prop("short_description", "string"),
					prop("full_description", "string"),
					prop("opportunity_type", "string"),
					prop("vacancy_level", "string"),
					prop("employment_type", "string"),
					prop("work_format", "string"),
					prop("location_id", "string"),
				),
				"CuratorCreateInput": objectSchema(
					prop("email", "string"),
					prop("password", "string"),
					prop("display_name", "string"),
					prop("curator_type", "string"),
				),
				"StatusPayload": objectSchema(prop("status", "string"), prop("comment", "string")),
			},
		},
		"paths": buildPaths(),
	}
	return json.MarshalIndent(spec, "", "  ")
}

func prop(name, typ string) map[string]any {
	return map[string]any{"name": name, "schema": map[string]any{"type": typ}}
}

func objectSchema(props ...map[string]any) map[string]any {
	properties := map[string]any{}
	for _, p := range props {
		properties[p["name"].(string)] = p["schema"]
	}
	return map[string]any{"type": "object", "properties": properties}
}

func bodyRef(ref string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/" + ref},
			},
		},
	}
}

func okResponse(description string) map[string]any {
	return map[string]any{"description": description}
}

func pathIDParam(name string) map[string]any {
	return map[string]any{"in": "path", "name": name, "required": true, "schema": map[string]any{"type": "string"}}
}

func authOperation(summary string) map[string]any {
	return map[string]any{"security": []map[string]any{{"UserHeader": []string{}}}, "summary": summary}
}

func buildPaths() map[string]any {
	return map[string]any{
		"/api/health": map[string]any{"get": map[string]any{"tags": []string{"health"}, "summary": "Проверка доступности сервиса", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/auth/register": map[string]any{"post": map[string]any{"tags": []string{"auth"}, "summary": "Регистрация пользователя", "requestBody": bodyRef("RegisterInput"), "responses": map[string]any{"201": okResponse("Created"), "400": okResponse("Bad request")}}},
		"/api/auth/login": map[string]any{"post": map[string]any{"tags": []string{"auth"}, "summary": "Авторизация пользователя", "requestBody": bodyRef("LoginInput"), "responses": map[string]any{"200": okResponse("OK"), "401": okResponse("Unauthorized")}}},
		"/api/auth/curator/login": map[string]any{"post": map[string]any{"tags": []string{"auth"}, "summary": "Авторизация куратора", "requestBody": bodyRef("LoginInput"), "responses": map[string]any{"200": okResponse("OK"), "401": okResponse("Unauthorized")}}},
		"/api/opportunities": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Список возможностей", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/opportunities/map": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Маркеры возможностей для карты", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/opportunities/{id}": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Карточка возможности", "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK"), "404": okResponse("Not found")}}},
		"/api/opportunities/{id}/applications": map[string]any{"post": merge(authOperation("Отклик на возможность"), map[string]any{"tags": []string{"public"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("ApplicationInput"), "responses": map[string]any{"201": okResponse("Created"), "400": okResponse("Bad request")}})},
		"/api/companies": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Список компаний", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/companies/{id}": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Карточка компании", "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK"), "404": okResponse("Not found")}}},
		"/api/tags": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Список тегов", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/cities": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Список городов", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/locations": map[string]any{"get": map[string]any{"tags": []string{"public"}, "summary": "Список локаций", "responses": map[string]any{"200": okResponse("OK")}}},
		"/api/me/student-profile": map[string]any{
			"get": merge(authOperation("Профиль соискателя"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"put": merge(authOperation("Обновить профиль соискателя"), map[string]any{"tags": []string{"student"}, "requestBody": bodyRef("StudentProfileInput"), "responses": map[string]any{"200": okResponse("OK")}}),
		},
		"/api/me/resumes": map[string]any{
			"get": merge(authOperation("Список резюме"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"post": merge(authOperation("Создать резюме"), map[string]any{"tags": []string{"student"}, "requestBody": bodyRef("ResumeInput"), "responses": map[string]any{"201": okResponse("Created")}}),
		},
		"/api/me/resumes/{id}/primary": map[string]any{"patch": merge(authOperation("Сделать резюме основным"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/portfolio-projects": map[string]any{
			"get": merge(authOperation("Список проектов портфолио"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"post": merge(authOperation("Создать проект портфолио"), map[string]any{"tags": []string{"student"}, "requestBody": bodyRef("PortfolioProjectInput"), "responses": map[string]any{"201": okResponse("Created")}}),
		},
		"/api/me/applications": map[string]any{"get": merge(authOperation("Мои отклики"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/favorite-opportunities": map[string]any{"get": merge(authOperation("Избранные возможности"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/favorite-opportunities/{opportunityId}": map[string]any{
			"post": merge(authOperation("Добавить возможность в избранное"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("opportunityId")}, "responses": map[string]any{"200": okResponse("OK")}}),
			"delete": merge(authOperation("Удалить возможность из избранного"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("opportunityId")}, "responses": map[string]any{"200": okResponse("OK")}}),
		},
		"/api/me/favorite-companies": map[string]any{"get": merge(authOperation("Избранные компании"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/favorite-companies/{companyId}": map[string]any{
			"post": merge(authOperation("Добавить компанию в избранное"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("companyId")}, "responses": map[string]any{"200": okResponse("OK")}}),
			"delete": merge(authOperation("Удалить компанию из избранного"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("companyId")}, "responses": map[string]any{"200": okResponse("OK")}}),
		},
		"/api/me/contacts": map[string]any{"get": merge(authOperation("Список контактов"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/contact-requests": map[string]any{
			"get": merge(authOperation("Список запросов в контакты"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"post": merge(authOperation("Создать запрос в контакты"), map[string]any{"tags": []string{"student"}, "requestBody": bodyRef("ContactRequestInput"), "responses": map[string]any{"201": okResponse("Created")}}),
		},
		"/api/me/contact-requests/{id}": map[string]any{"patch": merge(authOperation("Изменить статус запроса в контакты"), map[string]any{"tags": []string{"student"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("ContactRequestInput"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/me/recommendations": map[string]any{"post": merge(authOperation("Рекомендовать возможность контакту"), map[string]any{"tags": []string{"student"}, "requestBody": bodyRef("RecommendationInput"), "responses": map[string]any{"201": okResponse("Created")}})},
		"/api/me/notifications": map[string]any{"get": merge(authOperation("Мои уведомления"), map[string]any{"tags": []string{"student"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/employer/company": map[string]any{
			"get": merge(authOperation("Профиль компании работодателя"), map[string]any{"tags": []string{"employer"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"put": merge(authOperation("Обновить профиль компании"), map[string]any{"tags": []string{"employer"}, "requestBody": bodyRef("CompanyInput"), "responses": map[string]any{"200": okResponse("OK")}}),
		},
		"/api/employer/company-links": map[string]any{"post": merge(authOperation("Добавить ссылку компании"), map[string]any{"tags": []string{"employer"}, "requestBody": bodyRef("CompanyLinkInput"), "responses": map[string]any{"201": okResponse("Created")}})},
		"/api/employer/company-verifications": map[string]any{"post": merge(authOperation("Отправить компанию на верификацию"), map[string]any{"tags": []string{"employer"}, "requestBody": bodyRef("VerificationInput"), "responses": map[string]any{"201": okResponse("Created")}})},
		"/api/employer/opportunities": map[string]any{
			"get": merge(authOperation("Список возможностей работодателя"), map[string]any{"tags": []string{"employer"}, "responses": map[string]any{"200": okResponse("OK")}}),
			"post": merge(authOperation("Создать возможность"), map[string]any{"tags": []string{"employer"}, "requestBody": bodyRef("OpportunityInput"), "responses": map[string]any{"201": okResponse("Created")}}),
		},
		"/api/employer/opportunities/{id}": map[string]any{
			"get": merge(authOperation("Получить возможность работодателя"), map[string]any{"tags": []string{"employer"}, "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK")}}),
			"patch": merge(authOperation("Обновить возможность работодателя"), map[string]any{"tags": []string{"employer"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("OpportunityInput"), "responses": map[string]any{"200": okResponse("OK")}}),
		},
		"/api/employer/opportunities/{id}/applications": map[string]any{"get": merge(authOperation("Список откликов на возможность"), map[string]any{"tags": []string{"employer"}, "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/employer/applications/{id}/status": map[string]any{"patch": merge(authOperation("Изменить статус отклика"), map[string]any{"tags": []string{"employer"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StatusPayload"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/users": map[string]any{"post": merge(authOperation("Создать куратора"), map[string]any{"tags": []string{"curator"}, "requestBody": bodyRef("CuratorCreateInput"), "responses": map[string]any{"201": okResponse("Created")}})},
		"/api/curator/users/{id}/status": map[string]any{"patch": merge(authOperation("Изменить статус пользователя"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StatusPayload"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/student-profiles/{id}": map[string]any{"patch": merge(authOperation("Обновить профиль соискателя"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StudentProfileInput"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/employer-profiles/{id}": map[string]any{"patch": merge(authOperation("Обновить профиль работодателя"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/moderation-queue": map[string]any{"get": merge(authOperation("Очередь модерации"), map[string]any{"tags": []string{"curator"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/moderation-queue/{id}": map[string]any{"patch": merge(authOperation("Решение по элементу модерации"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StatusPayload"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/company-verifications": map[string]any{"get": merge(authOperation("Список верификаций компаний"), map[string]any{"tags": []string{"curator"}, "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/company-verifications/{id}": map[string]any{"patch": merge(authOperation("Решение по верификации компании"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StatusPayload"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/opportunities/{id}/status": map[string]any{"patch": merge(authOperation("Изменить статус возможности"), map[string]any{"tags": []string{"curator"}, "parameters": []any{pathIDParam("id")}, "requestBody": bodyRef("StatusPayload"), "responses": map[string]any{"200": okResponse("OK")}})},
		"/api/curator/audit-logs": map[string]any{"get": merge(authOperation("Журнал аудита"), map[string]any{"tags": []string{"curator"}, "responses": map[string]any{"200": okResponse("OK")}})},
	}
}

func merge(base map[string]any, extra map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

func SwaggerUIHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Tramplin API Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/doc.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
    });
  </script>
</body>
</html>`
}
