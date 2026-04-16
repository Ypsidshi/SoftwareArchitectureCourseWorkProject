¬Тестирование белым ящиком



Листинг 1 – Функция проверки промокода



func (c Code) Validate() error {

&#x09;if len(c) != codeLen {

&#x09;	return ErrInvalidCodeFormat

&#x09;}

&#x09;for \_, ch := range string(c) {

&#x09;	if !((ch >= 'a' \&\& ch <= 'z') || (ch >= '0' \&\& ch <= '9')) {

&#x09;		return ErrInvalidCodeFormat

&#x09;	}

&#x09;}

&#x09;return nil

}



"Картинка"







Часть 1. Покрытие операторов



№	Операторы	Входные данные	Выходные данные

1	d	abc123	ErrInvalidCodeFormat

2	k, j	abcde\_1234	ErrInvalidCodeFormat









Часть 2. Покрытие решений



№	Решения	Входные данные	Выходные данные

1	c - да	abc123	ErrInvalidCodeFormat

2	c – нет, g – да, нет, i - нет	abc1234567	nil

3	c – нет, g – да, i - да	/bc1234567	ErrInvalidCodeFormat







Часть 3. Покрытие условий



№	Условие	Входные данные	Выходные данные

1	c - да	abc123	ErrInvalidCodeFormat

2	c – нет, g – да, нет, i1 – да, нет, i2 – да,  i3 – да,  i4 – да	abc1234567	nil

3	c – нет, g – да, нет, i1 – да, i2 – нет,  i3 – да,  i4 – нет	{bc1234567	ErrInvalidCodeFormat

4	c – нет, g – да, i1 – нет,  i3 – нет	/bc1234567	ErrInvalidCodeFormat



Часть 4. Покрытие решений и условий



№	Условие	Входные данные	Выходные данные

1	c - да	abc123	ErrInvalidCodeFormat

2	c – нет, g – да, нет, i1 – да, нет, i2 – да,  i3 – да,  i4 – да	abc1234567	nil

3	c – нет, g – да, нет, i1 – да, i2 – нет,  i3 – да,  i4 – нет	{bc1234567	ErrInvalidCodeFormat

4	c – нет, g – да, i1 – нет,  i3 – нет	/bc1234567	ErrInvalidCodeFormat







Часть 5. Комбинаторное покрытие условий



1\. с = да

2\. с = нет

3\. g = да

4\. g = нет

5\. i1 = нет, i2 = да, i3 = нет, i4 = да

6\. i1 = нет, i2 = да, i3 = да, i4 = да

7\. i1 = нет, i2 = да, i3 = да, i4 = нет

8\. i1 = да, i2 = да, i3 = да, i4 = нет

9\. i1 = да, i2 = нет, i3 = да, i4 = нет



№	Условие	Входные данные	Выходные данные

1	c – да	abc123	ErrInvalidCodeFormat

2	c – нет, g – да, нет, i1 – да, нет, i2 – да,  i3 – да,  i4 – да	abc1234567	nil

3	c – нет, g – да, i1 – да, i2 – нет,  i3 – не вычисляется,  i4 –  не вычисляется	{bc1234567	ErrInvalidCodeFormat

4	c – нет, g – да, i1 – нет, i2 – не вычисляется, i3 - нет, i4 – не вычисляется	/bc1234567	ErrInvalidCodeFormat

5	c – нет, g – да, i1 – нет, i2 – не вычисляется, i3 – да, i4 - нет	:bc1234567	ErrInvalidCodeFormat





Часть 6. Управляющий граф программы



"Картинка графа"



1 – проверка len(code) != codeLen

2 – res = ErrInvalidCodeFormat

3 – i = 0

4 – i < len(code)

5 – return res

6 – ch = code\[i]

7 – проверка символа

8 – res = ErrInvalidCodeFormat

9 – i = i + 1





