package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, address, status, created_at) VALUES (?, ?, ?, ?)`,
		p.Client,
		p.Address,
		p.Status,
		p.CreatedAt,
	)
	if err != nil {
		// если при добавлении произошла ошибка — возвращаем 0 и ошибку
		return 0, err
	}

	// спрашиваем у базы: "какой номер у только что добавленной посылки?"
	id, err := res.LastInsertId()
	if err != nil {
		// если не смогли получить номер — возвращаем 0 и ошибку
		return 0, err
	}

	// возвращаем номер посылки (преобразуем к int) и nil как "ошибки нет"
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRow(
		`SELECT number,client,status,address,created_at FROM parcel WHERE number = ?`, number)

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	// если ошибка — возвращаем пустую посылку и ошибку

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query(
		`SELECT number,client,status,address, created_at FROM parcel WHERE client = ?`, client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		var p Parcel

		err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`, status, number)
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	query := `
	 UPDATE parcel
	 SET address = ?
	 WHERE number = ? AND status = ?
	 `
	result, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		// Если произошла ошибка при выполнении запроса — возвращаем её.
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("нельзя изменить адрес: посылка не найдена или её статус не 'registered'")
	}
	// Всё прошло успешно.
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// SQL-запрос: удаляем строку только если статус = registered
	query := `
        DELETE FROM parcel
        WHERE number = ? AND status = ?
    `

	// Выполняем запрос
	result, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}

	// Узнаём, сколько строк реально удалилось
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Если 0 — значит:
	// - либо нет такой посылки
	// - либо статус не registered
	if rowsAffected == 0 {
		return fmt.Errorf("нельзя удалить: посылка не найдена или её статус не 'registered'")
	}

	// Всё прошло успешно
	return nil
}
