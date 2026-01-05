ALTER TABLE bot_configs
ADD COLUMN after_continue_text text
DEFAULT 'Отправь текст, голос, фото или документ для урока.';

ALTER TABLE bot_configs
ADD COLUMN no_voice_minutes_text text;