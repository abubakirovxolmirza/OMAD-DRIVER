-- Add location columns to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS from_latitude DECIMAL(10, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS from_longitude DECIMAL(11, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS from_address TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS to_latitude DECIMAL(10, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS to_longitude DECIMAL(11, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS to_address TEXT;

-- Clear existing data
TRUNCATE TABLE districts CASCADE;
TRUNCATE TABLE regions CASCADE;

-- Insert all regions of Uzbekistan
INSERT INTO regions (id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(1, 'Toshkent shahri', 'Тошкент шаҳри', 'Город Ташкент'),
(2, 'Toshkent viloyati', 'Тошкент вилояти', 'Ташкентская область'),
(3, 'Andijon', 'Андижон', 'Андижанская область'),
(4, 'Buxoro', 'Бухоро', 'Бухарская область'),
(5, 'Farg''ona', 'Фарғона', 'Ферганская область'),
(6, 'Jizzax', 'Жиззах', 'Джизакская область'),
(7, 'Xorazm', 'Хоразм', 'Хорезмская область'),
(8, 'Namangan', 'Наманган', 'Наманганская область'),
(9, 'Navoiy', 'Навоий', 'Навоийская область'),
(10, 'Qashqadaryo', 'Қашқадарё', 'Кашкадарьинская область'),
(11, 'Qoraqalpog''iston', 'Қорақалпоғистон', 'Республика Каракалпакстан'),
(12, 'Samarqand', 'Самарқанд', 'Самаркандская область'),
(13, 'Sirdaryo', 'Сирдарё', 'Сырдарьинская область'),
(14, 'Surxondaryo', 'Сурхондарё', 'Сурхандарьинская область');

-- Toshkent shahri tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(1, 'Bektemir', 'Бектемир', 'Бектемирский'),
(1, 'Chilonzor', 'Чилонзор', 'Чиланзарский'),
(1, 'Yashnobod', 'Яшнобод', 'Яшнабадский'),
(1, 'Mirobod', 'Миробод', 'Мирабадский'),
(1, 'Mirzo Ulug''bek', 'Мирзо Улуғбек', 'Мирзо-Улугбекский'),
(1, 'Olmazor', 'Олмазор', 'Алмазарский'),
(1, 'Sergeli', 'Сергели', 'Сергелийский'),
(1, 'Shayxontohur', 'Шайхонтоҳур', 'Шайхантахурский'),
(1, 'Uchtepa', 'Учтепа', 'Учтепинский'),
(1, 'Yakkasaroy', 'Яккасарой', 'Яккасарайский'),
(1, 'Yunusobod', 'Юнусобод', 'Юнусабадский');

-- Toshkent viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(2, 'Angren', 'Ангрен', 'Ангрен'),
(2, 'Bekobod', 'Бекобод', 'Бекабад'),
(2, 'Bo''ka', 'Бўка', 'Бука'),
(2, 'Bo''stonliq', 'Бўстонлиқ', 'Бостанлык'),
(2, 'Chinoz', 'Чиноз', 'Чиназ'),
(2, 'Qibray', 'Қибрай', 'Кибрай'),
(2, 'Nurafshon', 'Нурафшон', 'Нурафшан'),
(2, 'Ohangaron', 'Оҳангарон', 'Ахангаран'),
(2, 'Oqqo''rg''on', 'Оққўрғон', 'Аккурган'),
(2, 'Parkent', 'Паркент', 'Паркент'),
(2, 'Piskent', 'Пискент', 'Пскент'),
(2, 'Quyichirchiq', 'Қуйичирчиқ', 'Куйичирчик'),
(2, 'Toshkent', 'Тошкент', 'Ташкентский'),
(2, 'O''rta Chirchiq', 'Ўрта Чирчиқ', 'Уртачирчик'),
(2, 'Yangiyo''l', 'Янгийўл', 'Янгиюль'),
(2, 'Yuqorichirchiq', 'Юқоричирчиқ', 'Юкоричирчик'),
(2, 'Zangiota', 'Зангиота', 'Зангиата');

-- Andijon viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(3, 'Andijon shahri', 'Андижон шаҳри', 'Город Андижан'),
(3, 'Xonobod shahri', 'Хонобод шаҳри', 'Город Ханабад'),
(3, 'Andijon', 'Андижон', 'Андижанский'),
(3, 'Asaka', 'Асака', 'Асакинский'),
(3, 'Baliqchi', 'Балиқчи', 'Балыкчинский'),
(3, 'Bo''z', 'Бўз', 'Бузский'),
(3, 'Buloqboshi', 'Булоқбоши', 'Булакбашинский'),
(3, 'Jalaquduq', 'Жалақудуқ', 'Джалакудукский'),
(3, 'Izboskan', 'Избоскан', 'Избасканский'),
(3, 'Qo''rg''ontepa', 'Қўрғонтепа', 'Кургантепинский'),
(3, 'Marhamat', 'Марҳамат', 'Мархаматский'),
(3, 'Oltinko''l', 'Олтинкўл', 'Алтынкульский'),
(3, 'Paxtaobod', 'Пахтаобод', 'Пахтаабадский'),
(3, 'Shahrixon', 'Шаҳрихон', 'Шахриханский'),
(3, 'Ulug''nor', 'Улуғнор', 'Улугнорский'),
(3, 'Xo''jaobod', 'Хўжаобод', 'Ходжаабадский');

-- Buxoro viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(4, 'Buxoro shahri', 'Бухоро шаҳри', 'Город Бухара'),
(4, 'Buxoro', 'Бухоро', 'Бухарский'),
(4, 'G''ijduvon', 'Ғиждувон', 'Гиждуванский'),
(4, 'Jondor', 'Жондор', 'Жондорский'),
(4, 'Kogon', 'Когон', 'Каганский'),
(4, 'Olot', 'Олот', 'Алатский'),
(4, 'Peshku', 'Пешку', 'Пешкунский'),
(4, 'Qorako''l', 'Қоракўл', 'Каракульский'),
(4, 'Qorovulbozor', 'Қоровулбозор', 'Караулбазарский'),
(4, 'Romitan', 'Ромитан', 'Ромитанский'),
(4, 'Shofirkon', 'Шофиркон', 'Шафирканский'),
(4, 'Vobkent', 'Вобкент', 'Вабкентский');

-- Farg'ona viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(5, 'Farg''ona shahri', 'Фарғона шаҳри', 'Город Фергана'),
(5, 'Marg''ilon shahri', 'Марғилон шаҳри', 'Город Маргилан'),
(5, 'Quvasoy shahri', 'Қувасой шаҳри', 'Город Кувасай'),
(5, 'Qo''qon shahri', 'Қўқон шаҳри', 'Город Коканд'),
(5, 'Oltiariq', 'Олтиариқ', 'Алтыарыкский'),
(5, 'Bag''dod', 'Бағдод', 'Багдадский'),
(5, 'Beshariq', 'Бешариқ', 'Бешарыкский'),
(5, 'Buvayda', 'Бувайда', 'Бувайдинский'),
(5, 'Dang''ara', 'Данғара', 'Дангаринский'),
(5, 'Farg''ona', 'Фарғона', 'Ферганский'),
(5, 'Furqat', 'Фурқат', 'Фуркатский'),
(5, 'O''zbekiston', 'Ўзбекистон', 'Узбекистанский'),
(5, 'Qo''shtepa', 'Қўштепа', 'Куштепинский'),
(5, 'Quva', 'Қува', 'Кувинский'),
(5, 'Rishton', 'Ришton', 'Риштанский'),
(5, 'So''x', 'Сўх', 'Сохский'),
(5, 'Toshloq', 'Тошлоқ', 'Ташлакский'),
(5, 'Uchko''prik', 'Учкўприк', 'Учкуприкский'),
(5, 'Yozyovon', 'Ёзёвон', 'Язъяванский');

-- Jizzax viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(6, 'Jizzax shahri', 'Жиззах шаҳри', 'Город Джизак'),
(6, 'Arnasoy', 'Арнасой', 'Арнасайский'),
(6, 'Baxmal', 'Бахмал', 'Бахмальский'),
(6, 'Do''stlik', 'Дўстлик', 'Дустликский'),
(6, 'Forish', 'Фориш', 'Форишский'),
(6, 'G''allaorol', 'Ғаллаорол', 'Галляаральский'),
(6, 'Sharof Rashidov', 'Шароф Рашидов', 'Шараф-Рашидовский'),
(6, 'Mirzacho''l', 'Мирзачўл', 'Мирзачульский'),
(6, 'Paxtakor', 'Пахтакор', 'Пахтакорский'),
(6, 'Yangiobod', 'Янгиобод', 'Янгиабадский'),
(6, 'Zafarobod', 'Зафаробод', 'Зафарабадский'),
(6, 'Zomin', 'Зомин', 'Заминский');

-- Xorazm viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(7, 'Urganch shahri', 'Урганч шаҳри', 'Город Ургенч'),
(7, 'Xiva shahri', 'Хива шаҳри', 'Город Хива'),
(7, 'Bog''ot', 'Боғот', 'Багатский'),
(7, 'Gurlan', 'Гурлан', 'Гурленский'),
(7, 'Xonqa', 'Хонқа', 'Ханкинский'),
(7, 'Hazorasp', 'Ҳазорасп', 'Хазараспский'),
(7, 'Qo''shko''pir', 'Қўшкўпир', 'Кушкупырский'),
(7, 'Shovot', 'Шовот', 'Шаватский'),
(7, 'Urganch', 'Урганч', 'Ургенчский'),
(7, 'Xiva', 'Хива', 'Хивинский'),
(7, 'Yangiariq', 'Янгиариқ', 'Янгиарыкский'),
(7, 'Yangibozor', 'Янгибозор', 'Янгибазарский');

-- Namangan viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(8, 'Namangan shahri', 'Наманган шаҳри', 'Город Наманган'),
(8, 'Chortoq', 'Чортоқ', 'Чартакский'),
(8, 'Chust', 'Чуст', 'Чустский'),
(8, 'Kosonsoy', 'Косонсой', 'Касансайский'),
(8, 'Mingbuloq', 'Мингбулоқ', 'Мингбулакский'),
(8, 'Namangan', 'Наманган', 'Наманганский'),
(8, 'Norin', 'Норин', 'Нарынский'),
(8, 'Pop', 'Поп', 'Папский'),
(8, 'To''raqo''rg''on', 'Тўрақўрғон', 'Туракурганский'),
(8, 'Uchqo''rg''on', 'Учқўрғон', 'Учкурганский'),
(8, 'Uychi', 'Уйчи', 'Уйчинский'),
(8, 'Yangiqo''rg''on', 'Янгиқўрғон', 'Янгикурганский');

-- Navoiy viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(9, 'Navoiy shahri', 'Навоий шаҳри', 'Город Навои'),
(9, 'Zarafshon shahri', 'Зарафшон шаҳри', 'Город Зарафшан'),
(9, 'Karmana', 'Кармана', 'Карманинский'),
(9, 'Konimex', 'Конимех', 'Канимехский'),
(9, 'Navbahor', 'Навбаҳор', 'Навбахорский'),
(9, 'Nurota', 'Нурота', 'Нуратинский'),
(9, 'Qiziltepa', 'Қизилтепа', 'Кызылтепинский'),
(9, 'Tomdi', 'Томди', 'Тамдынский'),
(9, 'Uchquduq', 'Учқудуқ', 'Учкудукский'),
(9, 'Xatirchi', 'Хатирчи', 'Хатырчинский');

-- Qashqadaryo viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(10, 'Qarshi shahri', 'Қарши шаҳри', 'Город Карши'),
(10, 'Chiroqchi', 'Чироқчи', 'Чиракчинский'),
(10, 'Dehqonobod', 'Деҳқонобод', 'Дехканабадский'),
(10, 'G''uzor', 'Ғузор', 'Гузарский'),
(10, 'Kasbi', 'Касби', 'Касбийский'),
(10, 'Kitob', 'Китоб', 'Китабский'),
(10, 'Koson', 'Косон', 'Касанский'),
(10, 'Mirishkor', 'Миришкор', 'Миришкорский'),
(10, 'Muborak', 'Муборак', 'Мубарекский'),
(10, 'Nishon', 'Нишон', 'Нишанский'),
(10, 'Qamashi', 'Қамаши', 'Камашинский'),
(10, 'Qarshi', 'Қарши', 'Каршинский'),
(10, 'Shahrisabz', 'Шаҳрисабз', 'Шахрисабзский'),
(10, 'Yakkabog''', 'Яккабоғ', 'Яккабагский');

-- Qoraqalpog'iston tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(11, 'Nukus shahri', 'Нукус шаҳри', 'Город Нукус'),
(11, 'Amudaryo', 'Амударё', 'Амударьинский'),
(11, 'Beruniy', 'Беруний', 'Берунийский'),
(11, 'Chimboy', 'Чимбой', 'Чимбайский'),
(11, 'Ellikqal''a', 'Элликқалъа', 'Элликкалинский'),
(11, 'Kegeyli', 'Кегейли', 'Кегейлийский'),
(11, 'Mo''ynoq', 'Мўйноқ', 'Муйнакский'),
(11, 'Nukus', 'Нукус', 'Нукусский'),
(11, 'Qanliko''l', 'Қанликўл', 'Канлыкульский'),
(11, 'Qorao''zak', 'Қораўзак', 'Караузякский'),
(11, 'Qo''ng''irot', 'Қўнғирот', 'Кунградский'),
(11, 'Shumanay', 'Шуманай', 'Шуманайский'),
(11, 'Taxtako''pir', 'Тахтакўпир', 'Тахтакупырский'),
(11, 'To''rtko''l', 'Тўрткўл', 'Турткульский'),
(11, 'Xo''jayli', 'Хўжайли', 'Ходжейлийский');

-- Samarqand viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(12, 'Samarqand shahri', 'Самарқанд шаҳри', 'Город Самарканд'),
(12, 'Kattaqo''rg''on shahri', 'Каттақўрғон шаҳри', 'Город Каттакурган'),
(12, 'Bulung''ur', 'Булунғур', 'Булунгурский'),
(12, 'Ishtixon', 'Иштихон', 'Иштыханский'),
(12, 'Jomboy', 'Жомбой', 'Джамбайский'),
(12, 'Kattaqo''rg''on', 'Каттақўрғон', 'Каттакурганский'),
(12, 'Narpay', 'Нарпай', 'Нарпайский'),
(12, 'Nurobod', 'Нуробод', 'Нурабадский'),
(12, 'Oqdaryo', 'Оқдарё', 'Акдарьинский'),
(12, 'Paxtachi', 'Пахтачи', 'Пахтачинский'),
(12, 'Payariq', 'Паяриқ', 'Пайарыкский'),
(12, 'Pastdarg''om', 'Пастдарғом', 'Пастдаргомский'),
(12, 'Qo''shrabot', 'Қўшработ', 'Кошрабатский'),
(12, 'Samarqand', 'Самарқанд', 'Самаркандский'),
(12, 'Toyloq', 'Тойлоқ', 'Тайлакский'),
(12, 'Urgut', 'Ургут', 'Ургутский');

-- Sirdaryo viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(13, 'Guliston shahri', 'Гулистон шаҳри', 'Город Гулистан'),
(13, 'Akaltyn', 'Акалтын', 'Акалтынский'),
(13, 'Boyovut', 'Боёвут', 'Баяутский'),
(13, 'Guliston', 'Гулистон', 'Гулистанский'),
(13, 'Mirzaobod', 'Мирзаобод', 'Мирзаабадский'),
(13, 'Oqoltin', 'Оқолтин', 'Акалтынский'),
(13, 'Sardoba', 'Сардоба', 'Сардобский'),
(13, 'Sayhunobod', 'Сайхунобод', 'Сайхунабадский'),
(13, 'Sirdaryo', 'Сирдарё', 'Сырдарьинский'),
(13, 'Xovos', 'Ховос', 'Хавастский');

-- Surxondaryo viloyati tumanlari
INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru) VALUES
(14, 'Termiz shahri', 'Термиз шаҳри', 'Город Термез'),
(14, 'Angor', 'Ангор', 'Ангорский'),
(14, 'Bandixon', 'Бандихон', 'Бандыханский'),
(14, 'Boysun', 'Бойсун', 'Байсунский'),
(14, 'Denov', 'Денов', 'Денауский'),
(14, 'Jarqo''rg''on', 'Жарқўрғон', 'Джаркурганский'),
(14, 'Qiziriq', 'Қизириқ', 'Кизирикский'),
(14, 'Qo''mqo''rg''on', 'Қўмқўрғон', 'Кумкурганский'),
(14, 'Muzrabot', 'Музработ', 'Музрабатский'),
(14, 'Oltinsoy', 'Олтинсой', 'Алтынсайский'),
(14, 'Sariosiyo', 'Сариосиё', 'Сариасийский'),
(14, 'Sherobod', 'Шеробод', 'Шерабадский'),
(14, 'Sho''rchi', 'Шўрчи', 'Шурчинский'),
(14, 'Termiz', 'Термиз', 'Термезский'),
(14, 'Uzun', 'Узун', 'Узунский');

-- Reset sequences
SELECT setval('regions_id_seq', (SELECT MAX(id) FROM regions));
SELECT setval('districts_id_seq', (SELECT MAX(id) FROM districts));
