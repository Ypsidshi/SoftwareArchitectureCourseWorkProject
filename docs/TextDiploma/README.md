# Пояснительная записка диплома (ВКР)

Каталог для текста ВКР и диаграмм. Сданная курсовая — `docs/TextCw/` (не редактировать).

**Навигация по файлам:** `DiplomaCurrent.md`

## Структура каталога

| Путь | Содержание |
|:-----|:-----------|
| `Section_02_03/` | Разделы 2–3: `Section.md`, `Tables.md`, соответствующие `.docx` |
| `Section_04/` | Раздел 4 «Тестирование»: `Testing.md`, `Testing.docx` |
| `Appendices/` | Приложения А–Л: `Appendices.md`, `Appendix_B_Specification.md` |
| `Diagrams/` | PlantUML (`.puml`) и PNG для рисунков 1–21 |
| `Examples/` | Образцы текста: `Section_02_03/`, `Section_04/`, `Appendices/` |
| `fixes.md` | Журнал правок |
| `reference-template.dotx` | Опционально; для сборки с таблицами **не** использовать (см. ниже) |

## Связанные материалы

- Код: `src/go-microservices/`, `src/frontend/`
- SQL: `sql/postgres/`
- Справка: `docs/project-understanding-guide.md`, `docs/project-readiness-audit.md`

## Рендер диаграмм

```powershell
java -jar /c/Users/super/.vscode/extensions/jebbs.plantuml-2.18.1/plantuml.jar -tpng -charset UTF-8 -o c:/Users/super/source/repos/SoftwareArchitectureCourseWorkProject/docs/TextDiploma/Diagrams "c:/Users/super/source/repos/SoftwareArchitectureCourseWorkProject/docs/TextDiploma/Diagrams/*.puml"
```

## Сборка DOCX

Команды выполнять **из соответствующей папки**. Для файлов с **таблицами** флаг `--reference-doc` **не использовать** — иначе в Word ломается оформление таблиц.

**Разделы 2–3:**

```powershell
Set-Location docs/TextDiploma/Section_02_03
pandoc Section.md -o Section.docx --from markdown --to docx
pandoc Tables.md -o Tables.docx --from markdown --to docx
```

**Раздел 4:**

```powershell
Set-Location docs/TextDiploma/Section_04
pandoc Testing.md -o Testing.docx --from markdown --to docx
```

**Приложения:**

```powershell
Set-Location docs/TextDiploma/Appendices
pandoc Appendix_B_Specification.md -o Appendix_B_Specification.docx --from markdown --to docx
pandoc Appendices.md -o Appendices.docx --from markdown --to docx
```

Готовые фрагменты в Word правятся вручную; повторная сборка docx перезаписывает файл целиком.
