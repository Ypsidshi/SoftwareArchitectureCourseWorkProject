Пояснительная записка диплома — навигация

Рабочие материалы разнесены по папкам. Полный текст ПЗ в Word собирается из частей (разделы 1 и 5–7 дополняются отдельно).

## Карта каталога `docs/TextDiploma/`

```
TextDiploma/
├── DiplomaCurrent.md          ← этот файл (навигация)
├── README.md                  ← команды pandoc, PlantUML
├── fixes.md
├── reference-template.dotx
├── Section_02_03/             ← разделы 2 и 3
├── Section_04/                ← раздел 4 «Тестирование»
├── Appendices/                ← приложения А–Л
├── Diagrams/                  ← PlantUML и PNG (рис. 1–21)
├── Examples/                  ← образцы по разделам
└── Уже существующие наработки/
```

## Где что лежит

| Часть ПЗ | Папка | Основной файл Markdown | DOCX |
|:---------|:------|:------------------------|:-----|
| Разделы 2–3 (проектирование, реализация) | `Section_02_03/` | `Section.md` | `Section.docx` |
| Таблицы для Word (разд. 2–3) | `Section_02_03/` | `Tables.md` | `Tables.docx` |
| Раздел 4 (тестирование) | `Section_04/` | `Testing.md` | `Testing.docx` |
| Приложения В–Л, заглушки | `Appendices/` | `Appendices.md` | `Appendices.docx` |
| Приложение Б (спецификация БД) | `Appendices/` | `Appendix_B_Specification.md` | `Appendix_B_Specification.docx` |
| Диаграммы | `Diagrams/` | `*.puml` → `*.png` | — |

## Порядок сборки полной ПЗ в Word

1. Раздел 1 — `Уже существующие наработки/` или новый файл (пока нет в структуре).
2. Разделы 2–3 — `Section_02_03/Section.docx` (или вставка из `Section.md`).
3. Раздел 4 — `Section_04/Testing.docx`.
4. Разделы 5–7 — дополняются.
5. Список источников.
6. Приложения — `Appendices/Appendix_B_Specification.docx`, затем `Appendices/Appendices.docx`.

Таблицы из `Section_02_03/Tables.docx` удобно копировать в Word без стилей шаблона.

Подробные команды сборки — в `README.md`.
