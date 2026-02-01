API
API — программный интерфейс системы для взаимодействия с системами ТСП.

Интерфейс работает по адресу api.cloudpayments.ru и поддерживает функции для выполнения платежа, отмены оплаты, возврата денег, завершения платежей, выполненных по двухстадийной схеме, создания и отмены подписок на рекуррентные платежи, а также отправки счетов по почте.

Принцип работы
Параметры передаются методом POST в теле запроса в формате «ключ=значение» либо в JSON.
API может принимать не больше 150 000 полей в одном запросе. Тайм-аут на получение ответа от API — 5 минут.
Во всех запросах к API если передать число с дробной частью в целочисленное поле, то ошибки не будет, зато произойдёт математическое округление.
API ограничивает максимальное количество одновременных запросов для тестовых терминалов до 5, для боевых до 30. Если количество обрабатываемых в данный момент запросов к сайту больше ограничения - API будет возвращать ответ с HTTP кодом 429 (Too many Requests) до момента пока не будет завершена обработка хотя бы одного запроса. При необходимости пересмотра ограничений - обратитесь к персональному менеджеру.
Выбор формата передачи параметров определяется на стороне клиента и управляется через заголовок запроса Content-Type.

Для параметров «ключ=значение» Content-Type: application/x-www-form-urlencoded;
Для параметров JSON Content-Type: application/json;
Ответ система выдает в JSON-формате, который как минимум включает в себя два параметра: Success и Message:

{ "Success": false, "Message": "Invalid Amount value" }
Аутентификация запросов
Для аутентификации запроса используется HTTP Basic Auth — отправка логина и пароля в заголовке HTTP-запроса. В качестве логина используется Public ID, в качестве пароля — API Secret. Оба этих значения можно получить в личном кабинете.

Если в запросе не передан заголовок с данными аутентификации или переданы неверные данные, система вернет HTTP-статус 401 – Unauthorized.
API secret используется для обеспечения безопасности. Он должен храниться в защищенном месте.
Идемпотентность API
Идемпотентность — свойство API при повторном запросе выдавать тот же результат, что на первичный запрос без повторной обработки. Это значит, что вы можете отправить несколько запросов к системе с одинаковым идентификатором, при этом обработан будет только один успешный запрос, а все ответы будут идентичными. Таким образом реализуется защита от сетевых ошибок, которые приводят к созданию дублированных записей и действий.
Для включения идемпотентности необходимо в запросе к API передавать заголовок с ключом X-Request-ID, содержащий уникальный идентификатор. Формирование идентификатора запроса остается на вашей стороне — это может быть guid, комбинация из номера заказа, даты и суммы или любое другое значение на ваше усмотрение.
Каждый новый запрос, который необходимо обработать, должен включать новое значение X-Request-ID. Обработанный результат хранится в системе в течение 1 часа.

Тестовый метод
Для проверки взаимодействия с API можно вызвать тестовый метод.

Адрес метода:
https://api.cloudpayments.ru/test

Параметры запроса:
Отсутствуют.

Пример ответа:
В ответ метод возвращает статус запроса.

{"Success":true,"Message":"bd6353c3-0ed6-4a65-946f-083664bf8dbd"}
Оплата по криптограмме
По требованию НСПК, для ТСП, предоставляющих телекоммуникационные услуги (MCC 4814 Telecommunication Services), обязательна передача номера телефона плательщика. При отсутствии номера телефона в запросе, банки-эквайеры будут отклонять проводимые операции.

Номер телефона необходимо передавать в объекте Payer в параметре Phone.
Формат передачи: только цифры, без разделителей “( )“, “-“ .
Пример передачи: “+71234567890“
Метод для оплаты по криптограмме платежных данных результат алгоритма шифрования. Для формирования криптограммы воспользуйтесь скриптом Checkout.

Адреса метода:
https://api.cloudpayments.ru/payments/cards/charge — для одностадийного платежа
https://api.cloudpayments.ru/payments/cards/auth — для двухстадийного

Параметры запроса:

Параметр Формат Применение Описание
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Необязательный Валюта: RUB/USD/EUR/GBP (см. справочник). Если параметр не передан, то по умолчанию принимает значение RUB
IpAddress String Обязательный IP-адрес плательщика
CardCryptogramPacket String Обязательный Криптограмма платежных данных
Name String Необязательный Имя держателя карты латиницей
PaymentUrl String Необязательный Адрес сайта, с которого совершается вызов скрипта checkout
InvoiceId String Необязательный Номер счета или заказа
Description String Необязательный Описание оплаты в свободной форме
CultureName String Необязательный Язык уведомлений. Возможные значения: "ru-RU", "en-US". (см. справочник)
AccountId String Необязательный Обязательный идентификатор пользователя для создания подписки и получения токена
Email String Необязательный E-mail плательщика, на который будет отправлена квитанция об оплате
Payer Object Необязательный Доп. поле, куда передается информация о плательщике. Используйте следующие параметры: FirstName, LastName, MiddleName, Birth, Street, Address, City, Country, Phone, Postcode
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией, в том числе инструкции для создания подписки или формирования онлайн-чека должны обёртываться в объект cloudpayments. Мы зарезервировали названия следующих параметров и отображаем их содержимое в реестре операций, выгружаемом в Личном Кабинете: name, firstName, middleName, lastName, nick, phone, address, comment, birthDate.
SaveCard bool Необязательный Признак сохранения карточного токена для проведения оплаты по сохранённой карте (см. Оплата по токену (рекарринг)).
Возможные значения: true - после успешной оплаты будет возвращён карточный токен, false - токен не будет возвращаться (по-умолчанию)
Параметр SaveCard обрабатывается только при включении настройки "Сохранение токена карты" в Личном Кабинете. При включении настройки "Сохранять принудительно", параметр SaveCard будет игнорироваться.
Параметр amount не принимает сумму транзакции меньше 0.01.
В ответ сервер возвращает JSON с тремя составляющими:

поле success — результат запроса;
поле message — описание ошибки;
объект model — расширенная информация.
Возможные варианты ответа:

Некорректно сформирован запрос:
success — false
message — описание ошибки
Требуется 3-D Secure аутентификация:
success — false
model — информация для проведения аутентификации
Транзакция отклонена:
success — false
model — информация о транзакции и код ошибки
Транзакция принята:
success — true
model — информация о транзакции
Результат запроса (поле Success) не отражает статус транзакции, а лишь свидетельствует об успешности/неуспешности запроса API. Статус транзакции отражен в дополнительном поле Status, также его можно узнать в результате получения уведомлений. Интерпретацию статусов транзакции смотрите в справочнике.
Пример запроса на оплату по криптограмме:

{
"Amount":10,
"Currency":"RUB",
"InvoiceId":"1234567",
"IpAddress": "123.123.123.123",
"Description":"Оплата товаров в example.com",
"AccountId":"user_x",
"Name":"CARDHOLDER NAME", // CardCryptogramPacket Обязательный параметр
"CardCryptogramPacket":"01492500008719030128SMfLeYdKp5dSQVIiO5l6ZCJiPdel4uDjdFTTz1UnXY+3QaZcNOW8lmXg0H670MclS4lI+qLkujKF4pR5Ri+T/E04Ufq3t5ntMUVLuZ998DLm+OVHV7FxIGR7snckpg47A73v7/y88Q5dxxvVZtDVi0qCcJAiZrgKLyLCqypnMfhjsgCEPF6d4OMzkgNQiynZvKysI2q+xc9cL0+CMmQTUPytnxX52k9qLNZ55cnE8kuLvqSK+TOG7Fz03moGcVvbb9XTg1oTDL4pl9rgkG3XvvTJOwol3JDxL1i6x+VpaRxpLJg0Zd9/9xRJOBMGmwAxo8/xyvGuAj85sxLJL6fA=="
"Payer":
{
"FirstName":"Тест",
"LastName":"Тестов",
"MiddleName":"Тестович",
"Birth":"1955-02-24",
"Address":"тестовый проезд дом тест",
"Street":"Lenina",
"City":"MO",
"Country":"RU",
"Phone":"+71234567890",
"Postcode":"345"
}
}

Пример ответа: некорректный запрос:

{"Success":false,"Message":"Amount is required"}
Пример ответа: требуется 3-D Secure аутентификация:

{
"Model": {
"TransactionId": 891463508,
"PaReq": "+/eyJNZXJjaGFudE5hbWUiOm51bGwsIkZpcnN0U2l4IjoiNDI0MjQyIiwiTGFzdEZvdXIiOiI0MjQyIiwiQW1vdW50IjoxMDAuMCwiQ3VycmVuY3lDb2RlIjoiUlVCIiwiRGF0ZSI6IjIwMjEtMTAtMjVUMDA6MDA6MDArMDM6MDAiLCJDdXN0b21lck5hbWUiOm51bGwsIkN1bHR1cmVOYW1lIjoicnUtUlUifQ==",
"GoReq": null,
"AcsUrl": "https://demo.cloudpayments.ru/acs",
"ThreeDsSessionData": null,
"IFrameIsAllowed": true,
"FrameWidth": null,
"FrameHeight": null,
"ThreeDsCallbackId": "7be4d37f0a434c0a8a7fc0e328368d7d",
"EscrowAccumulationId": null
},
"Success": false,
"Message": null
}
Пример ответа: транзакция отклонена. В поле ReasonCode код ошибки (см. справочник):

{
"Model": {
"ReasonCode": 5051,
"PublicId": "pk\_****************\*\*****************",
"TerminalUrl": "http://test.test",
"TransactionId": 891583633,
"Amount": 10,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 10,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": "1234567",
"AccountId": "user_x",
"Email": null,
"Description": "Оплата товаров в example.com",
"JsonData": null,
"CreatedDate": "/Date(1635154784619)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-25T09:39:44",
"AuthDate": null,
"AuthDateIso": null,
"ConfirmDate": null,
"ConfirmDateIso": null,
"AuthCode": null,
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "123.123.123.123",
"IpCountry": "CN",
"IpCity": "Beijing",
"IpRegion": "Beijing",
"IpDistrict": "Beijing",
"IpLatitude": 39.9289,
"IpLongitude": 116.3883,
"CardFirstSix": "400005",
"CardLastFour": "5556",
"CardExpDate": "12/25",
"CardType": "Visa",
"CardProduct": null,
"CardCategory": null,
"EscrowAccumulationId": null,
"IssuerBankCountry": "US",
"Issuer": "ITS Bank",
"CardTypeCode": 0,
"Status": "Declined",
"StatusCode": 5,
"CultureName": "ru",
"Reason": "InsufficientFunds",
"CardHolderMessage": "Недостаточно средств на карте",
"Type": 0,
"Refunded": false,
"Name": "CARDHOLDER NAME",
"Token": "tk_255c42192323f2e09ea17635302c3",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": false,
"Message": null
}
Пример ответа: транзакция принята:

{
"Model": {
"ReasonCode": 0,
"PublicId": "pk\_**************\*\*\*\***************",
"TerminalUrl": "http://test.test",
"TransactionId": 891510444,
"Amount": 10,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 10,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": "1234567",
"AccountId": "user_x",
"Email": null,
"Description": "Оплата товаров в example.com",
"JsonData": null,
"CreatedDate": "/Date(1635150224630)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-25T08:23:44",
"AuthDate": "/Date(1635150224739)/",
"AuthDateIso": "2021-10-25T08:23:44",
"ConfirmDate": null,
"ConfirmDateIso": null,
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "123.123.123.123",
"IpCountry": "CN",
"IpCity": "Beijing",
"IpRegion": "Beijing",
"IpDistrict": "Beijing",
"IpLatitude": 39.9289,
"IpLongitude": 116.3883,
"CardFirstSix": "411111",
"CardLastFour": "1111",
"CardExpDate": "11/25",
"CardType": "Visa",
"CardProduct": "C",
"CardCategory": "Visa Signature (Signature)",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "CloudPayments",
"CardTypeCode": 0,
"Status": "Authorized",
"StatusCode": 2,
"CultureName": "ru",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 0,
"Refunded": false,
"Name": "CARDHOLDER NAME",
"Token": "0a0afb77-8f41-4de2-9524-1057f9695303",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": true,
"Message": null
}
Обработка 3-D Secure
Для проведения 3-D Secure аутентификации нужно отправить плательщика на адрес, указанный в параметре AcsUrl ответа сервера с передачей следующих параметров:

MD — параметр TransactionId из ответа сервера;
PaReq — одноименный параметр из ответа сервера;
TermUrl — адрес на вашем сайте для возврата плательщика после аутентификации.
Регистр букв в названии параметров имеет значение.
Пример формы:

<form name="downloadForm" action="AcsUrl" method="POST">
    <input type="hidden" name="PaReq" value="eJxVUdtugkAQ/RXDe9mLgo0Z1nhpU9PQasWmPhLYAKksuEChfn13uVR9mGTO7MzZM2dg3qSn0Q+X\nRZIJxyAmNkZcBFmYiMgxDt7zw6MxZ+DFkvP1ngeV5AxcXhR+xEdJ6BhpEZnEYLBdfPAzg56JKSKT\nAhqgGpFB7IuSgR+cl5s3NqFTG2NAPYSUy82aETqeWPYUUAdB+ClnwSmrwtz/TbkoC0BtDYKsEqX8\nZfZkDGgAUMkTi8synyFU17V5N2nKCpBuAHRVs610VijCJgmZu17UXTxhFWP34l7evYPlegsHkO6A\n0C85o5hMsI3piNIZHc+IBaitg59qJYzgdrUOQK7/WNy+3FZAeSqV5cMqAwLe5JlQwpny8T8HdFW8\netFuBqUyahV+Hjf27vWCaSx22fe+KY6kXKZfJLK1x22TZkyUS8QiHaUGgDQN6s+H+tOq7O7kf8hd\nt30=">
    <input type="hidden" name="MD" value="504">
    <input type="hidden" name="TermUrl" value="https://example.com/post3ds?order=1234567">
</form>
<script>
    window.onload = submitForm;
    function submitForm() { downloadForm.submit(); }
</script>
После аутентификации плательщик будет возвращен на TermUrl с параметрами MD и PaRes, переданными методом POST.

Для завершения оплаты выполните следующий метод Post3ds.

Адрес метода:
https://api.cloudpayments.ru/payments/cards/post3ds

Параметры запроса:

Параметр Формат Применение Описание
TransactionId Long Обязательный Значение параметра MD
PaRes String Обязательный Значение одноименного параметра
В ответ на корректно сформированный запрос сервер вернет либо информацию об успешной транзакции, либо — об отклоненной.

Оплата по токену (рекарринг)
По требованию НСПК, для ТСП, предоставляющих телекоммуникационные услуги (MCC 4814 Telecommunication Services), обязательна передача номера телефона плательщика. При отсутствии номера телефона в запросе, банки-эквайеры будут отклонять проводимые операции.

Номер телефона необходимо передавать в объекте Payer в параметре Phone.
Формат передачи: только цифры, без разделителей “( )“, “-“ .
Пример передачи: “+71234567890“
Метод для оплаты по токену, полученному при оплате по криптограмме, либо через Pay-уведомление.

Оплата по токену проходит только на том терминале (publicId), на котором был получен токен. Если необходимо использовать токены на других терминалах, обратитесь в свободной форме на почту assistant@cp.ru
Финальный статус оплаты по токену СБП (“Завершена”/“Отклонена”) можно получить через Pay и Fail-уведомления уже после создания транзакции, а не в ответе на сам запрос оплаты по токену.
Адреса метода:
https://api.cloudpayments.ru/payments/tokens/charge — для одностадийного платежа
https://api.cloudpayments.ru/payments/tokens/auth — для двухстадийного

Параметры запроса:

Параметр Формат Применение Описание
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Необязательный Валюта: RUB/USD/EUR/GBP (см. справочник). Если параметр не передан, то по умолчанию принимает значение RUB
AccountId String Обязательный Идентификатор пользователя
TrInitiatorCode Int Обязательный Признак инициатора списания денежных средств.
Возможные значения:
0 - транзакция инициирована ТСП на основе ранее сохраненных учетных данных;
1 - транзакция инициирована держателем карты (клиентом) на основе ранее сохраненных учетных данных.
ВАЖНО! В случае, если транзакция инициирована ТСП, необходимо дополнительно в запросе указать параметр PaymentScheduled с корректным значением.
PaymentScheduled Int Необязательный Признак оплаты по расписанию на основе ранее сохраненных учетных данных.
Возможные значения:
0 - без расписания;
1 - по расписанию.
ВАЖНО! В случае, если при запросе данный параметр не указан, по умолчанию будет использоваться значение 0.
Token String Обязательный Токен
InvoiceId String Необязательный Номер счета или заказа
Description String Необязательный Назначение платежа в свободной форме
IpAddress String Необязательный IP-адрес плательщика
Email String Необязательный E-mail плательщика, на который будет отправлена квитанция об оплате
Payer Object Необязательный Доп. поле, куда передается информация о плательщике. Используйте следующие параметры: FirstName, LastName, MiddleName, Birth, Street, Address, City, Country, Phone, Postcode
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией, в том числе инструкции для создания подписки или формирования онлайн-чека должны обёртываться в объект cloudpayments. Мы зарезервировали названия следующих параметров и отображаем их содержимое в реестре операций, выгружаемом в Личном Кабинете: name, firstName, middleName, lastName, nick, phone, address, comment, birthDate.
В ответ сервер возвращает JSON с тремя составляющими: поле success — результат запроса, поле message — описание ошибки, объект model — расширенная информация.

Возможные варианты
Некорректно сформирован запрос:
success — false
message — описание ошибки
Транзакции отклонена:
success — false
model — информация о транзакции и код ошибки
Транзакции принята:
success — true
model — информация о транзакции
Результат запроса (поле Success) не отражает статус транзакции, а лишь свидетельствует об успешности/неуспешности запроса API. Статус транзакции отражен в дополнительном поле Status, также его можно узнать в результате получения уведомлений. Интерпретацию статусов транзакции смотрите в справочнике.
Пример запроса на оплату по токену:

{
"Amount":59,
"Currency":"RUB",
"InvoiceId":"1234567",
"Description":"Оплата товаров в example.com",
"AccountId":"user_x",
"TrInitiatorCode": 0,
"PaymentScheduled": 1,
"Token":"success_1111a3e0-2428-48fb-a530-12815d90d0e8",
"Payer":
{
"FirstName":"Тест",
"LastName":"Тестов",
"MiddleName":"Тестович",
"Birth":"1955-02-24",
"Address":"тестовый проезд дом тест",
"Street":"Lenina",
"City":"MO",
"Country":"RU",
"Phone":"+71234567890",
"Postcode":"345"
}
}
Пример ответа: некорректный запрос

{"Success":false,"Message":"Amount is required"}
Пример ответа: транзакция отклонена. В поле ReasonCode код ошибки (см. справочник)

{
"Model": {
"ReasonCode": 5051,
"PublicId": "pk\_****************\*\*****************",
"TerminalUrl": "http://test.test",
"TransactionId": 891583633,
"Amount": 100,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 100,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": "1234567",
"AccountId": "user_x",
"TrInitiatorCode": 0,
"Email": null,
"Description": "Оплата товаров в example.com",
"JsonData": null,
"CreatedDate": "/Date(1635154784619)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-25T09:39:44",
"AuthDate": null,
"AuthDateIso": null,
"ConfirmDate": null,
"ConfirmDateIso": null,
"AuthCode": null,
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "123.123.123.123",
"IpCountry": "CN",
"IpCity": "Beijing",
"IpRegion": "Beijing",
"IpDistrict": "Beijing",
"IpLatitude": 39.9289,
"IpLongitude": 116.3883,
"CardFirstSix": "400005",
"CardLastFour": "5556",
"CardExpDate": "12/25",
"CardType": "Visa",
"CardProduct": null,
"CardCategory": null,
"EscrowAccumulationId": null,
"IssuerBankCountry": "US",
"Issuer": "ITS Bank",
"CardTypeCode": 0,
"Status": "Declined",
"StatusCode": 5,
"CultureName": "ru",
"Reason": "InsufficientFunds",
"CardHolderMessage": "Недостаточно средств на карте",
"Type": 0,
"Refunded": false,
"Name": "CARDHOLDER NAME",
"Token": "tk_255c42192323f2e09ea17635302c3",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": false,
"Message": null
}
Пример ответа: транзакция принята

{
"Model": {
"ReasonCode": 0,
"PublicId": "pk\_************\*\*************",
"TerminalUrl": "http://test.test",
"TransactionId": 897728064,
"Amount": 59,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 59,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": 1234567,
"AccountId": "user_x",
"TrInitiatorCode": 0,
"Email": null,
"Description": "Оплата товаров в example.com",
"JsonData": null,
"CreatedDate": "/Date(1635562705992)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-30T02:58:25",
"AuthDate": "/Date(1635562706070)/",
"AuthDateIso": "2021-10-30T02:58:26",
"ConfirmDate": "/Date(1635562706070)/",
"ConfirmDateIso": "2021-10-30T02:58:26",
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "87.251.91.164",
"IpCountry": "RU",
"IpCity": "Новосибирск",
"IpRegion": "Новосибирская область",
"IpDistrict": "Сибирский федеральный округ",
"IpLatitude": 55.03923,
"IpLongitude": 82.927818,
"CardFirstSix": "424242",
"CardLastFour": "4242",
"CardExpDate": "12/25",
"CardType": "Visa",
"CardProduct": "I",
"CardCategory": "Visa Infinite Infinite",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "CloudPayments",
"CardTypeCode": 0,
"Status": "Completed",
"StatusCode": 3,
"CultureName": "ru-RU",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 0,
"Refunded": false,
"Name": "SER",
"Token": "success_1111a3e0-2428-48fb-a530-12815d90d0e8",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": true,
"Message": null
}
Подтверждение оплаты
Для платежей, проведенных по двухстадийной схеме, необходимо подтверждение оплаты, которое можно выполнить через личный кабинет, либо через вызов метода API.

Адрес метода:
https://api.cloudpayments.ru/payments/confirm

Параметры запроса:

Параметр Формат Применение Описание
TransactionId Long Обязательный Номер транзакции в системе
Amount Number Обязательный Сумма подтверждения в валюте транзакции, разделитель точка. Количество не нулевых знаков после точки – 2.
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией, в том числе инструкции для формирования онлайн-чека
Пример запроса:

{"TransactionId":455,"Amount":65.98}
Пример ответа:

{"Success":true,"Message":null}
Отмена оплаты
Отмену оплаты можно выполнить через личный кабинет либо через вызов метода API.

Адрес метода:
https://api.cloudpayments.ru/payments/void

Параметры запроса:

Параметр Формат Применение Описание
TransactionId Long Обязательный Номер транзакции в системе
Пример запроса:

{"TransactionId":455}
Пример ответа:

{"Success":true,"Message":null}
Возврат денег
Возврат денег можно выполнить через личный кабинет или через вызов метода API.

Адрес метода:
https://api.cloudpayments.ru/payments/refund

Параметры запроса:

Параметр Формат Применение Описание
TransactionId Long Обязательный Номер транзакции оплаты
Amount Number Обязательный Сумма возврата в валюте транзакции, разделитель точка. Количество не нулевых знаков после точки – 2.
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией, в том числе инструкции для формирования онлайн-чека
Пример запроса:

{"TransactionId":455, "Amount": 100}
Пример ответа:

{
"Model": {
"TransactionId": 568
},
"Success": true,
"Message": null
}
Возвраты по транзакциям старше года не проводятся.
Выплата по криптограмме
Выплату по криптограмме можно осуществить через вызов метода API.

Адрес метода:
https://api.cloudpayments.ru/payments/cards/topup

Параметры запроса:

Параметр Формат Применение Описание
CardCryptogramPacket String Обязательный Криптограмма платежных данных
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Обязательный Валюта: RUB
Name String Необязательный Имя держателя карты латиницей
AccountId String Необязательный Идентификатор пользователя
Email String Необязательный E-mail плательщика, на который будет отправлена квитанция об оплате
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией. Мы зарезервировали названия следующих параметров и отображаем их содержимое в реестре операций, выгружаемом в Личном Кабинете: name, firstName, middleName, lastName, nick, phone, address, comment, birthDate.
InvoiceId String Необязательный Номер заказа в вашей системе
Description String Необязательный Описание оплаты в свободной форме
Payer Object Необязательный Доп. поле, куда передается информация о плательщике
Receiver Object Необязательный Доп. поле, куда передается информация о получателе
Для выплат на иностранные карты понадобится параметр Payer или Receiver в зависимости от терминала.
Пример запроса:

{
"Name":"CARDHOLDER NAME",
"CardCryptogramPacket":"01492500008719030128SMfLeYdKp5dSQVIiO5l6ZCJiPdel4uDjdFTTz1UnXY+3QaZcNOW8lmXg0H670MclS4lI+qLkujKF4pR5Ri+T/E04Ufq3t5ntMUVLuZ998DLm+OVHV7FxIGR7snckpg47A73v7/y88Q5dxxvVZtDVi0qCcJAiZrgKLyLCqypnMfhjsgCEPF6d4OMzkgNQiynZvKysI2q+xc9cL0+CMmQTUPytnxX52k9qLNZ55cnE8kuLvqSK+TOG7Fz03moGcVvbb9XTg1oTDL4pl9rgkG3XvvTJOwol3JDxL1i6x+VpaRxpLJg0Zd9/9xRJOBMGmwAxo8/xyvGuAj85sxLJL6fA==",
"Amount":1,
"AccountId":"user@example.com",
"Currency":"RUB",
"InvoiceId":"1234567",
"Payer":
//Набор полей аналогичен и для параметра Receiver
{
"FirstName":"Тест",
"LastName":"Тестов",
"MiddleName":"Тестович",
"Address":"тестовый проезд дом тест",
"Birth":"1955-02-24",
"City":"MO",
"Street":"Ленина",
"Country":"RU",
"Phone":"+71234567890",
"Postcode":"345"
}
}
Пример ответа:

{
"Model": {
"ReasonCode": 0,
"PublicId": "pk\_**************\*\*\***************",
"TerminalUrl": "http://test.test",
"TransactionId": 897739551,
"Amount": 399,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 399,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": null,
"AccountId": "user@example.com",
"Email": null,
"Description": null,
"JsonData": null,
"CreatedDate": "/Date(1635564719715)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-30T03:31:59",
"AuthDate": "/Date(1635564719858)/",
"AuthDateIso": "2021-10-30T03:31:59",
"ConfirmDate": "/Date(1635564719858)/",
"ConfirmDateIso": "2021-10-30T03:31:59",
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "172.18.200.124",
"IpCountry": "",
"IpCity": null,
"IpRegion": null,
"IpDistrict": null,
"IpLatitude": null,
"IpLongitude": null,
"CardFirstSix": "424242",
"CardLastFour": "4242",
"CardExpDate": "12/25",
"CardType": "Visa",
"CardProduct": "I",
"CardCategory": "Visa Infinite (Infinite)",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "CloudPayments",
"CardTypeCode": 0,
"Status": "Completed",
"StatusCode": 3,
"CultureName": "ru",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 2,
"TransactionIsInProcess": false,
"Refunded": false,
"Name": "CARDHOLDER NAME",
"Token": null,
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": true,
"Message": null
}
Можно воспользоваться механизмом надежной аутентификации запроса на выплату. Для этого передайте в нашу поддержку сертификат с публичной частью ключа. Далее сгенерируйте подпись на основе тела запроса и разместите ее в заголовке X-Signature в base64 формате. CloudPayments проверит вашу подпись, используя CryptoService. Если подпись валидна, то обработка запроса продолжится, если нет — обработка прекратится.
В ответе метода выплаты по криптограмме POST /payments/cards/topup возвращаются следующие статусы:

Статус Описание
Completed Выплата выполнена успешно
Created Выплата в промежуточном статусе
Failed Выплата не выполнена
В случае перехода транзакции в промежуточный статус срабатывает сценарий асинхронной обработки выплаты, в результате выполнения которого параметр Model.Status принимает значение “Created”, а флаг Model.TransactionIsInProcess становится равен true.

При завершении обработки выплаты статус транзакции будет изменен на терминальный и отправлено webhook-уведомление с финальным статусом:

Статус Тип уведомления
Completed PAY, Документация CloudPayments
Failed FAIL, Документация CloudPayments
В редких случаях процесс обработки запроса на выплату может занимать до 60 минут
Выплата по токену
Выплату по токену можно осуществить через вызов метода API.

Выплаты по токену можно сделать только по токену, который получили при оплате по карте.

Токены TPay, Sbp и другие токены, полученные на иных способах оплаты, не подходят для выплат по токену.

Адрес метода:
https://api.cloudpayments.ru/payments/token/topup

Параметры запроса:

Параметр Формат Применение Описание
Token String Обязательный Токен карты, выданный системой после первого платежа
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
AccountId String Обязательный Идентификатор пользователя
Currency String Обязательный Валюта: RUB
InvoiceId String Необязательный Номер заказа в вашей системе
Payer Object Необязательный Доп. поле, куда передается информация о плательщике
Receiver Object Необязательный Доп. поле, куда передается информация о получателе
Для выплат на иностранные карты понадобится параметр Payer или Receiver в зависимости от терминала.
Пример запроса:

{
"Token":"0a0afb77-8f41-4de2-9524-1057f9695303",
"Amount":59,
"AccountId":"user_x",
"Currency":"RUB",
"Payer":
//Набор полей аналогичен и для параметра Receiver
{
"FirstName":"Тест",
"LastName":"Тестов",
"MiddleName":"Тестович",
"Address":"тестовый проезд дом тест",
"Birth":"1955-02-24",
"City":"MO",
"Country":"RU",
"Phone":"+71234567890",
"Postcode":"345"
}
}
Пример ответа:

{
"Model": {
"ReasonCode": 0,
"PublicId": "pk\_**************\*\***************",
"TerminalUrl": "http://test.test",
"TransactionId": 897747761,
"Amount": 59,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 59,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": null,
"AccountId": "user_x",
"Email": null,
"Description": null,
"JsonData": null,
"CreatedDate": "/Date(1635566071122)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-30T03:54:31",
"AuthDate": "/Date(1635566071232)/",
"AuthDateIso": "2021-10-30T03:54:31",
"ConfirmDate": "/Date(1635566071232)/",
"ConfirmDateIso": "2021-10-30T03:54:31",
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "172.18.200.124",
"IpCountry": "",
"IpCity": null,
"IpRegion": null,
"IpDistrict": null,
"IpLatitude": null,
"IpLongitude": null,
"CardFirstSix": "411111",
"CardLastFour": "1111",
"CardExpDate": "11/25",
"CardType": "Visa",
"CardProduct": "C",
"CardCategory": "Visa Signature Signature",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "CloudPayments",
"CardTypeCode": 0,
"Status": "Completed",
"StatusCode": 3,
"CultureName": "ru",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 2,
"Refunded": false,
"Name": "CARDHOLDER NAME",
"Token": "0a0afb77-8f41-4de2-9524-1057f9695303",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": true,
"Message": null
}
Можно воспользоваться механизмом надежной аутентификации запроса на выплату. Для этого передайте в нашу поддержку сертификат с публичной частью ключа. Далее сгенерируйте подпись на основе тела запроса и разместите ее в заголовке X-Signature в base64 формате. CloudPayments проверит вашу подпись, используя CryptoService. Если подпись валидна, то обработка запроса продолжится, если нет — обработка прекратится.
Выплата по СБП
Выплату по СБП можно осуществить через вызов метода API.

Адрес метода:
https://api.cloudpayments.ru/payments/alt/topup

Параметры запроса:

Параметр Формат Применение Описание
MemberId String Необязательный Идентификатор банка получателя НПСК для СБП
Если в запросе поле Receiver.Phone заполнено, то MemberId обязательное
Допустимые значения для MemberId можно найти по ссылке. Значение из поля schema в поле MemberId надо передать только цифровое значение из поля schema
Например, для schema:bank100000000111 передать MemberId:100000000111
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Обязательный Валюта: RUB
AccountId String Необязательный Идентификатор пользователя
Email String Необязательный E-mail плательщика, на который будет отправлена квитанция об оплате
CultureName String Необязательный Язык уведомлений. Возможные значения: "ru-RU", "en-US". (см. справочник)
InvoiceId String Необязательный Номер заказа в вашей системе
Description String Необязательный Описание оплаты в свободной форме
Payer Object Необязательный Доп. поле, куда передается информация о плательщике. Используйте следующие параметры: FirstName, LastName, MiddleName, Birth, Street, Address, City, Country, Phone, Postcode
Receiver Object Обязательный Доп. поле, куда передается информация о получателе
Receiver.Phone String Обязательный Для выплаты по СБП необходимо указать номер получателя средств
Пример запроса:

{
"Amount":1,
"AccountId":"user@example.com",
"MemberId":"12456",
"Currency":"RUB",
"InvoiceId":"1234567",
"Receiver":
{
"Phone":"+71234567890"
}
}
Пример ответа:

{
"Model": {
"ReasonCode": 0,
"PublicId": "test_api_00000000000000000000002",
"TerminalUrl": "https://demo-preprod.cloudpayments.ru",
"TransactionId": 2200976211,
"Amount": 50,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 50,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": null,
"AccountId": "a.kadyrova@cloudpayments.ru",
"Email": "a.kadyrova@cloudpayments.ru",
"Description": null,
"JsonData": null,
"CreatedDate": "/Date(1717514966896)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2024-06-04T15:29:26",
"AuthDate": "/Date(1717514967037)/",
"AuthDateIso": "2024-06-04T15:29:27",
"ConfirmDate": "/Date(1717514967037)/",
"ConfirmDateIso": "2024-06-04T15:29:27",
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "172.18.200.35",
"IpCountry": "",
"IpCity": null,
"IpRegion": null,
"IpDistrict": null,
"IpLatitude": null,
"IpLongitude": null,
"CardFirstSix": "424242",
"CardLastFour": "4242",
"CardExpDate": "01/77",
"CardType": "Visa",
"CardProduct": "",
"CardCategory": "Не определен ()",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "TINKOFF",
"CardTypeCode": 0,
"Status": "Completed",
"StatusCode": 3,
"CultureName": "ru",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 2,
"Refunded": false,
"Name": "SERGEY",
"Token": null,
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0,
"IsLocalOrder": false,
"Gateway": 0,
"MasterPass": false,
"InfoShopData": null,
"Receiver":{
"FirstName":"Тест",
"LastName":"Тестов",
"MiddleName":"Тестович",
"Address":"тестовый проезд дом тест",
"Birth":"1955-02-24",
"City":"MO",
"Street":"Ленина",
"Country":"RU",
"Phone":"+71234567890",
"Postcode":"345"
}
},
"Success": true,
"Message": null,
"ErrorCode": null
}
Просмотр транзакции
Метод получения детализации по транзакции.

Адрес метода:
https://api.cloudpayments.ru/payments/get

Параметры запроса:

Параметр Формат Применение Описание
TransactionId Long Обязательный Номер транзакции
Если транзакция с указанным номером была найдена, система отобразит информацию о ней.

Пример запроса:

{"TransactionId":897749645}
Пример ответа:

{
"Model": {
"ReasonCode": 0,
"PublicId": "pk\_********************\*********************",
"TerminalUrl": "http://test.test",
"TransactionId": 897749645,
"Amount": 159,
"Currency": "RUB",
"CurrencyCode": 0,
"PaymentAmount": 159,
"PaymentCurrency": "RUB",
"PaymentCurrencyCode": 0,
"InvoiceId": "12345",
"AccountId": "usex_x",
"Email": null,
"Description": "test",
"JsonData": "",
"CreatedDate": "/Date(1635566398686)/",
"PayoutDate": null,
"PayoutDateIso": null,
"PayoutAmount": null,
"CreatedDateIso": "2021-10-30T03:59:58",
"AuthDate": "/Date(1635566402780)/",
"AuthDateIso": "2021-10-30T04:00:02",
"ConfirmDate": "/Date(1635566406382)/",
"ConfirmDateIso": "2021-10-30T04:00:06",
"AuthCode": "A1B2C3",
"TestMode": true,
"Rrn": null,
"OriginalTransactionId": null,
"FallBackScenarioDeclinedTransactionId": null,
"IpAddress": "87.251.91.164",
"IpCountry": "RU",
"IpCity": "Новосибирск",
"IpRegion": "Новосибирская область",
"IpDistrict": "Сибирский федеральный округ",
"IpLatitude": 55.03923,
"IpLongitude": 82.927818,
"CardFirstSix": "424242",
"CardLastFour": "4242",
"CardExpDate": "12/25",
"CardType": "Visa",
"CardProduct": "I",
"CardCategory": "Visa Infinite (Infinite)",
"EscrowAccumulationId": null,
"IssuerBankCountry": "RU",
"Issuer": "CloudPayments",
"CardTypeCode": 0,
"Status": "Completed",
"StatusCode": 3,
"CultureName": "ru-RU",
"Reason": "Approved",
"CardHolderMessage": "Оплата успешно проведена",
"Type": 0,
"Refunded": false,
"Name": "TEST",
"Token": "success_eb250528-bd9e-4de7-bb49-9e0b546351d3",
"SubscriptionId": null,
"GatewayName": "Test",
"AndroidPay": false,
"WalletType": "",
"TotalFee": 0
},
"Success": true,
"Message": null
}
Проверка статуса платежа
Метод поиска платежа и проверки статуса (см. справочник).

Адрес старого метода:
https://api.cloudpayments.ru/payments/find

Адрес нового метода:
https://api.cloudpayments.ru/v2/payments/find

Параметры запроса:

Параметр Формат Применение Описание
InvoiceId String Обязательный Номер заказа
Если платеж по указанному номеру заказа был найден, система отобразит либо информацию об успешной транзакции, либо — об отклоненной. Если будет найдено несколько платежей с указанным номером заказа, то система вернет информацию только о последней операции. Отличие нового метода в том, что он ищет по всем платежам, включая возвраты и выплаты на карту.

Пример запроса:

{"InvoiceId":"123456789"}
Пример ответа:

{"Success":false,"Message":"Not found"}
Проверка статуса платежа является избыточной и имеет смысл только в случае, если при проведении оплаты возник сбой, который привел к потере информации.
Выгрузка списка транзакций
Метод выгрузки списка транзакций за день.

Адрес метода:
https://api.cloudpayments.ru/payments/list

Параметры запроса:

Параметр Формат Применение Описание
Date Date Обязательный Дата создания операций
TimeZone String Необязательный Код временной зоны, по умолчанию — UTC
В выгрузку транзакций попадают все операции, зарегистрированные за указанный день. Для удобства учета вы можете указать код временной зоны (см. справочник).

Пример запроса:

{"Date":"2014-08-09", "TimeZone": "MSK"}
Пример ответа:

{
"Model":[
{
"PublicId":"test_api_00000000000000000000001",
"TerminalUrl":"https://cloudpayments.ru",
"TransactionId":54,
"Amount":12.34,
"Currency":"RUB",
"CurrencyCode":0,
"PaymentAmount":12.34,
"PaymentCurrency":"RUB",
"PaymentCurrencyCode":0,
"InvoiceId":"1234567",
"AccountId":"User@Example.com",
"Email":null,
"Description":"Оплата товаров в example.com",
"JsonData":"{\"some\": \"value\"}",
"CreatedDate":"\/Date(1615288374632)\/",
"PayoutDate":null,
"PayoutDateIso":null,
"PayoutAmount":null,
"CreatedDateIso":"2021-03-09T11:12:54",
"AuthDate":null,
"AuthDateIso":null,
"ConfirmDate":null,
"ConfirmDateIso":null,
"AuthCode":null,
"TestMode":true,
"Rrn":null,
"OriginalTransactionId":null,
"FallBackScenarioDeclinedTransactionId":null,
"IpAddress":"127.0.0.1",
"IpCountry":"",
"IpCity":null,
"IpRegion":null,
"IpDistrict":null,
"IpLatitude":null,
"IpLongitude":null,
"CardFirstSix":"424242",
"CardLastFour":"4242",
"CardExpDate":"05/22",
"CardType":"Visa",
"CardProduct":null,
"CardCategory":null,
"IssuerBankCountry":"FF",
"Issuer":null,
"CardTypeCode":0,
"Status":"Declined",
"StatusCode":5,
"CultureName":"ru",
"Reason":"SystemError",
"CardHolderMessage":"Повторите попытку позже",
"Type":0,
"Refunded":false,
"Name":"CARD HOLDER",
"Token":null,
"SubscriptionId":null,
"IsLocalOrder":false,
"HideInvoiceId":false,
"Gateway":0,
"GatewayName":"Test",
"AndroidPay":false,
"MasterPass":false,
"TotalFee":0,
"EscrowAccumulationId":null,
"ReasonCode":5096
}
],
"Success":true,
"Message":null
}
Выгрузка списка транзакций за произвольный период
Метод выгрузки списка транзакций за произвольный период.

Адрес метода:
https://api.cloudpayments.ru/v2/payments/list

Параметры запроса:

Параметр Формат Применение Описание
CreatedDateGte Date Обязательный Начальная дата создания операций
CreatedDateLte Date Обязательный Конечная дата создания операций
PageNumber Number Обязательный порядковый номер страницы, должно быть больше или равно 1
TimeZone String Необязательный Код временной зоны, по умолчанию — UTC
Statuses String Array Необязательный Статус операция. Может иметь значения [ "Authorized", "Completed", "Cancelled", "Declined" ]. По умолчанию выбраны все
В выгрузку транзакций попадают все операции, зарегистрированные за указанный день. Результат сортируется по CreatedDate, от старых к новым с лимитом - 100 на одну страницу. Для удобства учета вы можете указать код временной зоны (см. справочник).

Пример запроса:

{
"PageNumber": 1,
"CreatedDateGte":"2021-03-09T00:00:00+03:00",
"CreatedDateLte":"2021-03-10T00:00:00+03:00",
"TimeZone": "MSK",
"Statuses": [
"Authorized",
"Completed",
"Cancelled",
"Declined"
]
}
Пример ответа:

{
"Model":[
{
"PublicId":"test_api_00000000000000000000001",
"TerminalUrl":"https://cloudpayments.ru",
"TransactionId":54,
"Amount":12.34,
"Currency":"RUB",
"CurrencyCode":0,
"PaymentAmount":12.34,
"PaymentCurrency":"RUB",
"PaymentCurrencyCode":0,
"InvoiceId":"1234567",
"AccountId":"User@Example.com",
"Email":null,
"Description":"Оплата товаров в example.com",
"JsonData":"{\"some\": \"value\"}",
"CreatedDate":"\/Date(1615288374632)\/",
"PayoutDate":null,
"PayoutDateIso":null,
"PayoutAmount":null,
"CreatedDateIso":"2021-03-09T11:12:54",
"AuthDate":null,
"AuthDateIso":null,
"ConfirmDate":null,
"ConfirmDateIso":null,
"AuthCode":null,
"TestMode":true,
"Rrn":null,
"OriginalTransactionId":null,
"FallBackScenarioDeclinedTransactionId":null,
"IpAddress":"127.0.0.1",
"IpCountry":"",
"IpCity":null,
"IpRegion":null,
"IpDistrict":null,
"IpLatitude":null,
"IpLongitude":null,
"CardFirstSix":"424242",
"CardLastFour":"4242",
"CardExpDate":"05/22",
"CardType":"Visa",
"CardProduct":null,
"CardCategory":null,
"IssuerBankCountry":"FF",
"Issuer":null,
"CardTypeCode":0,
"Status":"Declined",
"StatusCode":5,
"CultureName":"ru",
"Reason":"SystemError",
"CardHolderMessage":"Повторите попытку позже",
"Type":0,
"Refunded":false,
"Name":"CARD HOLDER",
"Token":null,
"SubscriptionId":null,
"IsLocalOrder":false,
"HideInvoiceId":false,
"Gateway":0,
"GatewayName":"Test",
"AndroidPay":false,
"MasterPass":false,
"TotalFee":0,
"EscrowAccumulationId":null,
"ReasonCode":5096
}
],
"Success":true,
"Message":null
}
Выгрузка списка претензий за произвольный период
Метод выгрузки списка претензий за произвольный период, но не более чем за 1 год. Для выгрузки требуется отправить http-запрос с типом авторизации “Basic Auth”. Логин - public key (pk) терминала.

В выгрузку претензий попадают все претензий, зарегистрированные за указанный период. Результат сортируется по Date, от старых к новым с лимитом - 100 на одну страницу.

Адрес метода:
https://api.cloudpayments.ru/chargebacks/list

Параметры запроса:

Параметр Формат Применение Описание
CreatedDateGte Date Обязательный Начальная дата создания претензии
CreatedDateLte Date Обязательный Конечная дата создания претензии
PageNumber Number Обязательный Порядковый номер страницы, должно быть больше или равно 1
Пример запроса:

{
"PageNumber": 1,
"CreatedDateGte":"2021-03-09T00:00:00+03:00",
"CreatedDateLte":"2021-03-10T00:00:00+03:00"
}
Параметры ответа:

Параметр Формат Применение
Success bool Признак успешности запроса
Message string Описание ошибки в случае неуспеха
Model object Модель ответа с претензиями
ChargeBackId string Идентификатор претензии
TransactionId NumberLong Номер транзакции, по которой оформили претензию
TerminalUrl string Сайт
Gateway string Банк
Operation string Операция, значения:
Presentment - предъявление претензии
Representment - отзыв претензии
Type string Тип, значения:
Chargeback - претензия
Fee - штраф
Card string Маскированный номер карта
Date date Дата регистрации претензции в Банке
Amount decimal Сумма
Currency string Валюта
Пример ответа:

{
"Model": [
{
"ChargeBackId": "6811d4a17bab1c4b44c19371",
"TerminalUrl": "https://cloudpayments.ru/",
"TransactionId": 2200203594,
"Amount": 100,
"Currency": "RUB",
"Type": "Chargeback",
"Operation": "Presentment",
"Card": "520500*******3055",
"Gateway": "SberbankRu",
"Date": "/Date(1745960400000)/"
}
],
"Success": true,
"Message": null,
"ErrorCode": null
}
Выгрузка токенов
Метод выгрузки списка всех платежных токенов CloudPayments.

Адрес метода:
https://api.cloudpayments.ru/payments/tokens/list

Параметры запроса:

Параметр Формат Применение Описание
PageNumber Int Обязательный Порядковый номер страницы, должно быть больше или равно 1
Пример запроса:

{
"PageNumber": 1
}
Пример ответа:

{
"Model": [
{
"Token": "tk_020a924486aa4df254331afa33f2a",
"AccountId": "user_x",
"CardMask": "4242 42****** 4242",
"ExpirationDateMonth": 12,
"ExpirationDateYear": 2020
},
{
"Token": "tk_1a9f2f10253a30a7c5692a3fc4c17",
"AccountId": "user_x",
"CardMask": "5555 55****** 4444",
"ExpirationDateMonth": 12,
"ExpirationDateYear": 2021
},
{
"Token": "tk_b91062f0f2875909233ab66d0fc7b",
"AccountId": "user_x",
"CardMask": "4012 88****** 1881",
"ExpirationDateMonth": 12,
"ExpirationDateYear": 2022
}
],
"Success": true,
"Message": null
}
Создание подписки на рекуррентные платежи
Метод создания подписки на рекуррентные платежи.

Оплата по токену проходит только на том терминале (publicId), на котором был получен токен. Если необходимо использовать токены на других терминалах, обратитесь в свободной форме на почту assistant@cp.ru
Адрес метода:
https://api.cloudpayments.ru/subscriptions/create

Параметры запроса:

Параметр Формат Применение Описание
Token String Обязательный Токен карты, выданный системой после первого платежа
AccountId String Обязательный Идентификатор пользователя
Description String Обязательный Назначение платежа в свободной форме
Email String Необязательный E-mail плательщика
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Обязательный Валюта: RUB/USD/EUR/GBP (см. справочник)
RequireConfirmation Bool Обязательный Если значение true — платежи будут выполняться по двухстадийной схеме
StartDate DateTime Обязательный Дата и время первого платежа по плану во временной зоне UTC. Значение должно быть в будущем
Interval String Обязательный Интервал. Возможные значения: Day, Week, Month
Period Int Обязательный Период. В комбинации с интервалом, 1 Month значит раз в месяц, а 2 Week — раз в две недели. Должен быть больше 0
MaxPeriods Int Необязательный Максимальное количество платежей в подписке. Если указан, должен быть больше 0
CustomerReceipt json Необязательный Для изменения состава онлайн-чека
В ответ на корректно сформированный запрос система возвращает сообщение об успешно выполненной операции и идентификатор подписки.

Пример запроса:

{  
 "token": "477BBA133C182267FE5F086924ABDC5DB71F77BFC27F01F2843F2CDC69D89F05",
"accountId": "user_x",
"description": "Ежемесячная подписка на сервис example.com",
"email": "user@example.com",
"amount": 399,
"currency": "RUB",
"requireConfirmation": false,
"startDate": "2021-11-02T21:00:00",
"interval": "Day",
"period": 5
}
Пример ответа:

{
"Model": {
"Id": "sc_221da6421dc44dbd2cc3464f6f083",
"AccountId": "user_x",
"Description": "Ежемесячная подписка на сервис example.com",
"Email": "user@example.com",
"Amount": 399,
"CurrencyCode": 0,
"Currency": "RUB",
"RequireConfirmation": false,
"StartDate": "/Date(1635886800000)/",
"StartDateIso": "2021-11-02T21:00:00",
"IntervalCode": 2,
"Interval": "Day",
"Period": 5,
"MaxPeriods": null,
"CultureName": "ru-RU",
"StatusCode": 0,
"Status": "Active",
"SuccessfulTransactionsNumber": 0,
"FailedTransactionsNumber": 0,
"LastTransactionDate": null,
"LastTransactionDateIso": null,
"NextTransactionDate": "/Date(1635886800000)/",
"NextTransactionDateIso": "2021-11-02T21:00:00",
"Receipt": null,
"FailoverSchemeId": null
},
"Success": true,
"Message": null
}
Запрос информации о подписке
Метод получения информации о статусе подписки.

Адрес метода:
https://api.cloudpayments.ru/subscriptions/get

Параметры запроса:

Параметр Формат Применение Описание
Id String Обязательный Идентификатор подписки
Пример запроса:

{"Id":"sc_8cf8a9338fb8ebf7202b08d09c938"}
Пример ответа:

{
"Model": {
"Id": "sc_8cf8a9338fb8ebf7202b08d09c938",
"AccountId": "user_x",
"Description": null,
"Email": "user@example.com",
"Amount": 399,
"CurrencyCode": 0,
"Currency": "RUB",
"RequireConfirmation": false,
"StartDate": "/Date(1635886800000)/",
"StartDateIso": "2021-11-02T21:00:00",
"IntervalCode": 2,
"Interval": "Day",
"Period": 5,
"MaxPeriods": null,
"CultureName": "ru-RU",
"StatusCode": 3,
"Status": "Cancelled",
"SuccessfulTransactionsNumber": 0,
"FailedTransactionsNumber": 0,
"LastTransactionDate": null,
"LastTransactionDateIso": null,
"NextTransactionDate": null,
"NextTransactionDateIso": null,
"Receipt": null,
"FailoverSchemeId": null
},
"Success": true,
"Message": null
}
Поиск подписок
Метод получения списка подписок для определенного аккаунта.

Адрес метода:
https://api.cloudpayments.ru/subscriptions/find

Параметры запроса:

Параметр Формат Применение Описание
accountId String Обязательный Идентификатор пользователя
Пример запроса:

{"accountId":"user@example.com"}
Пример ответа:

{
"Model": [
{
"Id": "sc_221da6421dc44dbd2cc3464f6f083",
"AccountId": "user_x",
"Description": null,
"Email": "user@example.com",
"Amount": 399,
"CurrencyCode": 0,
"Currency": "RUB",
"RequireConfirmation": false,
"StartDate": "/Date(1635886800000)/",
"StartDateIso": "2021-11-02T21:00:00",
"IntervalCode": 2,
"Interval": "Day",
"Period": 5,
"MaxPeriods": null,
"CultureName": "ru-RU",
"StatusCode": 3,
"Status": "Cancelled",
"SuccessfulTransactionsNumber": 0,
"FailedTransactionsNumber": 0,
"LastTransactionDate": null,
"LastTransactionDateIso": null,
"NextTransactionDate": null,
"NextTransactionDateIso": null,
"Receipt": null,
"FailoverSchemeId": null
},
{
"Id": "sc_3ffc96c001e152b341817341b075a",
"AccountId": "user_x",
"Description": null,
"Email": "user@example.com",
"Amount": 999,
"CurrencyCode": 0,
"Currency": "RUB",
"RequireConfirmation": false,
"StartDate": "/Date(1635973200000)/",
"StartDateIso": "2021-11-03T21:00:00",
"IntervalCode": 2,
"Interval": "Day",
"Period": 5,
"MaxPeriods": null,
"CultureName": "ru-RU",
"StatusCode": 0,
"Status": "Active",
"SuccessfulTransactionsNumber": 0,
"FailedTransactionsNumber": 0,
"LastTransactionDate": null,
"LastTransactionDateIso": null,
"NextTransactionDate": "/Date(1635973200000)/",
"NextTransactionDateIso": "2021-11-03T21:00:00",
"Receipt": null,
"FailoverSchemeId": null
}
],
"Success": true,
"Message": null
}
Изменение подписки на рекуррентные платежи
Метод изменения ранее созданной подписки.

Адрес метода:
https://api.cloudpayments.ru/subscriptions/update

Параметры запроса:

Параметр Формат Применение Описание
Id String Обязательный Идентификатор подписки
Description String Необязательный Для изменения назначения платежа
Amount Number Необязательный Для изменения суммы платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2.
Currency String Необязательный Для изменения валюты: RUB/USD/EUR/GBP (см. справочник)
RequireConfirmation Bool Необязательный Для изменения схемы проведения платежей
StartDate DateTime Необязательный Для изменения даты и времени первого или следующего платежа во временной зоне UTC
Interval String Необязательный Для изменения интервала. Возможные значения: Day, Week, Month
Period Int Необязательный Для изменения периода. В комбинации с интервалом, 1 Month значит раз в месяц, а 2 Week — раз в две недели
MaxPeriods Int Необязательный Для изменения максимального количества платежей в подписке
CustomerReceipt json Необязательный Для изменения состава онлайн-чека
CultureName String Необязательный Язык уведомлений. Возможные значения: "ru-RU", "en-US". (см. справочник)
В случае завершения подписки или её отмены Плательщиком или Партнёром, при изменении любого атрибута подписки через API, подписка вновь активируется! В случае, если у подписки был задан MaxPeriods, то у вновь активированной подписки максимальное количество платежей в подписке будет равно MaxPeriods - количество платежей, которое уже было списано до момента отмены/завершения подписки.
В ответ на корректно сформированный запрос система возвращает сообщение об успешно выполненной операции и параметры подписки.

Пример запроса:

{  
 "Id":"sc_3ffc96c001e152b341817341b075a",
"description":"изменение рекуррента",
"amount":499,
"currency":"RUB"
}
Пример ответа:

{
"Model": {
"Id": "sc_3ffc96c001e152b341817341b075a",
"AccountId": "user_x",
"Description": "изменение рекуррента",
"Email": "user@example.com",
"Amount": 499,
"CurrencyCode": 0,
"Currency": "RUB",
"RequireConfirmation": false,
"StartDate": "/Date(1635973200000)/",
"StartDateIso": "2021-11-03T21:00:00",
"IntervalCode": 2,
"Interval": "Day",
"Period": 2,
"MaxPeriods": null,
"CultureName": "ru-RU",
"StatusCode": 0,
"Status": "Active",
"SuccessfulTransactionsNumber": 0,
"FailedTransactionsNumber": 0,
"LastTransactionDate": null,
"LastTransactionDateIso": null,
"NextTransactionDate": "/Date(1635973200000)/",
"NextTransactionDateIso": "2021-11-03T21:00:00",
"Receipt": null,
"FailoverSchemeId": null
},
"Success": true,
"Message": null
}
Отмена подписки на рекуррентные платежи
Метод отмены подписки на рекуррентные платежи.

Адрес метода:

https://api.cloudpayments.ru/subscriptions/cancel
Параметры запроса:

Параметр Формат Применение Описание
Id String Обязательный Идентификатор подписки
В ответ на корректно сформированный запрос система возвращает сообщение об успешно выполненной операции.

Пример запроса:

{"Id":"sc_cc673fdc50b3577e60eee9081e440"}
Пример ответа:

{"Success":true,"Message":null}
Вы также можете предоставить покупателю ссылку на сайт системы — https://my.cloudpayments.ru/, где он самостоятельно сможет найти и отменить свои регулярные платежи.

Создание счета для отправки по почте
Метод формирования ссылки на оплату и последующей отправки уведомления на e-mail адрес плательщика.

Адрес метода:
https://api.cloudpayments.ru/orders/create

Параметры запроса:

Параметр Формат Применение Описание
Amount Number Обязательный Cумма платежа в валюте, разделитель точка. Количество не нулевых знаков после точки – 2
Currency String Необязательный Валюта RUB/USD/EUR/GBP (см. справочник). Если параметр не передан, то по умолчанию принимает значение RUB
Description String Обязательный Назначение платежа в свободной форме
Email String Необязательный E-mail плательщика
RequireConfirmation Bool Необязательный Есть значение true — платеж будет выполнен по двухстадийной схеме
SendEmail Bool Необязательный Если значение true — плательщик получит письмо со ссылкой на оплату
InvoiceId String Необязательный Номер заказа в вашей системе
AccountId String Необязательный Идентификатор пользователя в вашей системе
OfferUri String Необязательный Ссылка на оферту, которая будет показываться на странице заказа
Phone String Необязательный Номер телефона плательщика в произвольном формате
SendSms Bool Необязательный Если значение true — плательщик получит СМС со ссылкой на оплату
SendViber Bool Необязательный Если значение true — плательщик получит сообщение в Viber со ссылкой на оплату
CultureName String Необязательный Язык уведомлений. Возможные значения: "ru-RU", "en-US". (см. справочник)
SubscriptionBehavior String Необязательный Для создания платежа с подпиской. Возможные значения: CreateWeekly, CreateMonthly
SuccessRedirectUrl String Необязательный Адрес страницы для редиректа при успешной оплате
FailRedirectUrl String Необязательный Адрес страницы для редиректа при неуспешной оплате
JsonData Json Необязательный Любые другие данные, которые будут связаны с транзакцией, в том числе инструкции для формирования онлайн-чека должны обёртываться в объект cloudpayments
В ответ на корректно сформированный запрос система возвращает параметры запроса и ссылку на оплату.
Пример запроса:

{
"Amount":10.0,
"Currency":"RUB",
"Description":"Оплата на сайте example.com",
"Email":"client@test.local",
"RequireConfirmation":true,
"SendEmail":false
}
Пример ответа:

{
"Model": {
"Id": "gASGZVgUN21hcpPF",
"Number": 2130,
"Amount": 10.0,
"Currency": "RUB",
"CurrencyCode": 0,
"Email": "client@test.local",
"Phone": "+71234567890",
"Description": "Оплата на сайте example.com",
"RequireConfirmation": true,
"Url": "https://orders.cloudpayments.ru/d/gASGZVgUN21hcpPF",
"CultureName": "ru-RU",
"CreatedDate": "/Date(1635835930555)/",
"CreatedDateIso": "2021-11-02T09:52:10",
"PaymentDate": null,
"PaymentDateIso": null,
"StatusCode": 0,
"Status": "Created",
"InternalId": 12272915
},
"Success": true,
"Message": null
}
Сообщение на телефон плательщика может быть отправлено только одним выбранным способом: СМС или Viber.
Отмена созданного счета
Метод отмены созданного счета:

Адрес метода:
https://api.cloudpayments.ru/orders/cancel

Параметры запроса:

Параметр Формат Применение Описание
Id String Обязательный Идентификатор счета
В ответ на корректно сформированный запрос система возвращает сообщение об успешно выполненной операции.

Пример запроса:

{"Id":"f2K8LV6reGE9WBFn"}
Пример ответа:

{
"Success": true,
"Message": null
}
Просмотр настроек уведомлений
Метод просмотра настроек уведомлений с указанием типа уведомления.

Адрес метода:
https://api.cloudpayments.ru/site/notifications/{Type}/get

Параметры запроса:

Параметр Формат Применение Описание
Type String Обязательный Тип уведомления: Check/Pay/Fail и т.д. (см. справочник)
Пример ответа на запрос для Pay-уведомления на адрес:
https://api.cloudpayments.ru/site/notifications/pay/get

{
"Model": {
"IsEnabled": true,
"Address": "http://example.com",
"HttpMethod": "POST",
"Encoding": "UTF8",
"Format": "CloudPayments"
},
"Success": true,
"Message": null
}
Изменение настроек уведомлений
Метод изменения настроек уведомлений.

Адрес метода:
https://api.cloudpayments.ru/site/notifications/{Type}/update

Параметры запроса:

Параметр Формат Применение Описание
Type String Обязательный Тип уведомления: Pay/Fail и т.д., кроме Check-уведомления (см. справочник)
IsEnabled Bool Необязательный Если значение true — то уведомление включено. Значение по умолчанию — false
Address String Необязательный, если IsEnabled=false, в противном случае обязательный Адрес для отправки уведомлений (для HTTPS-схемы необходим валидный SSL-сертификат)
HttpMethod String Необязательный HTTP-метод для отправки уведомлений. Возможные значения: GET, POST. Значение по умолчанию — GET
Encoding String Необязательный Кодировка уведомлений. Возможные значения: UTF8, Windows1251. Значение по умолчанию — UTF8
Format String Необязательный Формат уведомлений. Возможные значения: CloudPayments, QIWI, RT. Значение по умолчанию — CloudPayments
Пример запроса для Pay-уведомления на адрес:
https://api.cloudpayments.ru/site/notifications/pay/update:

{
"IsEnabled": true,
"Address": "http://example.com",
"HttpMethod": "GET",
"Encoding": "UTF8",
"Format": "CloudPayments"
}
Пример ответа:

{"Success":true,"Message":null}
Локализация
По умолчанию API выдает сообщения для пользователей на русском языке. Для получения ответов, локализованных для других языков, передайте в параметрах запроса CultureName.

Список поддерживаемых языков:

Язык Часовой пояс Код
Русский MSK ru-RU
Английский CET en-US
Латышский CET lv
Азербайджанский AZT az
Русский ALMT kk
Украинский EET uk
Польский CET pl
Вьетнамский ICT vi
Турецкий TRT tr
Длинная запись
