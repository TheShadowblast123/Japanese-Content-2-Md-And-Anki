from jisho_api.word import Word
from jisho_api.kanji import Kanji
import MeCab
import os
import glob
import re
import googletrans as gt
import string
import csv
from concurrent.futures import ThreadPoolExecutor, wait, ALL_COMPLETED
old_kanji = []
old_words = []
global content_md, kanji_md, sentences_md, words_md, content_path, kanji_path, sentences_path, words_path, csv_path
global current_name
current_name = ''
content_md = "Notes\Japanese Notes\Content.md"
kanji_md = "Notes\Japanese Notes\Kanji.md"
sentences_md = "Notes\Japanese Notes\Sentences.md"
words_md = "Notes\Japanese Notes\Words.md"
content_path = "Notes\Japanese Notes\Content"
kanji_path = "Notes\Japanese Notes\Kanji"
sentences_path = "Notes\Japanese Notes\Sentences"
words_path = "Notes\Japanese Notes\Words"
csv_path = r"Notes\Japanese Notes\CSV"

global skip_sentences
skip_sentences = False
kanji_range_1, kanji_range_2, kanji_range_3 = (0x3400, 0x4DBf),(0x4E00,0x9FCB), (0xF900, 0xFA6A)
kanji_set_1, kanji_set_2, kanji_set_3 = [chr(c) for c in range(*kanji_range_1)], [chr(c) for c in range(*kanji_range_2)], [chr(c) for c in range(*kanji_range_3)]
kanji_set = kanji_set_1 + kanji_set_2 + kanji_set_3

def parser(item) -> list[str]:
    mecab = MeCab.Tagger("-O wakati")
    return mecab.parse(item).split()

def check_title(title : str, test : str) -> bool:
    return title == f'[[{test}]]\n'

def intake_content() -> dict:
    output = {}
    new_content_path = "./New Content"

    txt_files = glob.glob(os.path.join(new_content_path, "*.txt"))
    if txt_files == []:
        print(f'You have no new sources place .txt files in {new_content_path} in order to begin')
        return
    
    for txt_file in txt_files:
        name = replace_spaces(os.path.basename(txt_file)).strip('.txt')
        print(name)
        with open(txt_file, "r", encoding='UTF8') as file:
            lines = file.readlines()
            file.close()
        blob = ''
        for line in lines:
            if line not in blob:
                temp = re.sub(r'[A-Za-z0-9]', '', line).replace(' ', '')
                if len(temp) > 0:
                    blob += temp
            continue

        output[name] = blob
        this_content_md = content_path + f'/{name}.md'

        
        with open(this_content_md, 'w', encoding='utf8') as file:
            file.writelines(lines)
            file.close()
    return output

def get_sentences() -> dict:
    punctuation = ['\n', '.', '?', '!', ' 〪', '。', ' 〭', '！', '．', '？']
    sources = intake_content()
    output = {}

    for n in list(sources.keys()):
        output[n] = []
    for name, content in sources.items():

        sentence = ''
        for c in content:
            if c not in punctuation:
                sentence += c
            else:
                if sentence:
                    if sentence not in output[name]:
                        output[name].append(sentence)
                    sentence = ''
    for name, content in sources.items():
        if name not in output:
            output[name] = []
        if sentence:
            output[name].append(sentence)
    return output

def sentence_to_word_string(sentence : str) -> str:
    word_punctuation = string.punctuation + r'！＂”“＃＄％＆＇（）＊＋，－．／：；＜＝＞？＠［＼］＾＿｀｛｜｝～、。〃〄々〆〇〈〉《》「」『』【】〒〓〔〕〖〗〘〙〚〛〜〝〞〟〠〡〢〣〤〥〦〧〨〩〪〭〮〯〫〬〰〱〲〳〴〵〶〷〸〹〺〻〼〽〾｟｠｡｢｣､･〿'
    pattern = f'[{re.escape(word_punctuation)}]'
    temp = str(parser(sentence))
    words_string = re.sub(pattern, '', temp)
    words_string = words_string.split()
    temp_array = []
    for word in words_string:
        current_length = len(temp_array)
        for c in word:
            if c in kanji_set or c in old_kanji:
                temp_array.append(f'[[Words/{word}|{word}]]')
                break
        if current_length == len(temp_array):
            temp_array.append(f'[[{word}]]')
    
    words_string = ' '.join(temp_array)

    return words_string

def word_to_kanji_string(word : str) -> str:
    temp = [ f'[[Kanji/{x}|{x}]]' if x in kanji_set or x in old_kanji else x for x in word]
    kanji = ''.join(temp)
    return kanji

def replace_spaces(tag : str) -> str:
    return tag.replace(' ', '_')

def append_content(name : str):
    with open(content_md, 'a+') as file:
        x = f'[[{name}]]\n'
        file.write(x)
        file.close()
        #we assume the content is new no matter what
    return

def append_sentence(sentence : str) -> bool:
    with open(sentences_md, 'r+', encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, sentence):
                file.close()
                return True
        file.close() 
    return False

def append_kanji(kanji : str) -> bool:
    with open(kanji_md, 'r+', encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, kanji):
                file.close()
                return True    
        file.close()
    return False

def add_new_stuff(kl : list[str], wl : list[str], sl : list[str]):
    temp = []
    with open(kanji_md, 'r', encoding="utf-8") as file:
        temp = file.readlines()
        temp.extend(kl)
        file.close()
    with open(kanji_md, 'w', encoding="utf-8") as file:
        file.writelines(temp)
        file.close()
    temp = []
    with open(words_md, 'r', encoding="utf-8") as file:
        temp = file.readlines()
        temp.extend(wl)
        file.close()
    with open(words_md, 'w', encoding="utf-8") as file:
        file.writelines(temp)
        file.close()
    temp = []
    with open(sentences_md, 'r', encoding="utf-8") as file:
        temp = file.readlines()
        temp.extend(sl)
        file.close()
    with open(sentences_md, 'w', encoding="utf-8") as file:
        file.writelines(temp)
        file.close()
    return

def append_word(word : str) -> bool:
    with open(words_md, 'r+', encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, word):
                file.close()
                return True
        file.close() 
    return False

def debugger(a):
    breakpoint()
    print(a)
    return

def edit_kanji_tags(item : str):
    lines = []
    with open (f'{kanji_path}\{item}.md', 'r', encoding='utf8') as file:
        lines = file.readlines()

        for line in lines:
            if 'Tags: ' in line:

                lines[lines.index(line)] = line[:-2] + ' ' f'[[{current_name}]] \n'
                break
            continue
        file.close()
    with open (f'{kanji_path}\{item}.md', 'w', encoding='utf8') as file:
        file.writelines(lines)
        file.close
    return

def edit_sentence_tags(item : str):
    lines = []
    with open (f'{sentences_path}\{item}.md', 'r', encoding='utf8') as file:
        lines = file.readlines()

        for line in lines:
            if 'Tags: ' in line:

                lines[lines.index(line)] = line[:-2] + ' ' f'[[{current_name}]] \n'
                break
            continue
        file.close()
    with open (f'{sentences_path}\{item}.md', 'w', encoding='utf8') as file:
        file.writelines(lines)
        file.close
    return

def edit_words_tags(item : str):
    lines = []
    with open (f'{words_path}\{item}.md', 'r', encoding='utf8') as file:
        lines = file.readlines()

        for line in lines:
            if 'Tags: ' in line:

                lines[lines.index(line)] = line[:-2] + ' ' f'[[{current_name}]] \n'
                break
            continue
        file.close()
    with open (f'{words_path}\{item}.md', 'w', encoding='utf8') as file:
        file.writelines(lines)
        file.close
    return

def write_to_kanji(l : list[str]) -> list[str]:
    edit_kanji_list = []
    count = 0
    copy = list(l)
    for k in copy:
        if append_kanji(k):
            edit_kanji_list.append(l.pop(count))
            
            continue
        count += 1
    return edit_kanji_list

def write_to_words(l : list[str]) -> list[str]:
    count = 0
    edit_words_list = []
    copy = list(l)
    for k in copy:
        if append_word(k):
            edit_words_list.append(l.pop(count))
            continue
        count += 1
    return edit_words_list

def write_to_sentences(l : list[str]) -> list[str]:
    edit_sentences_list = []
    count = 0
    copy = list(l)
    for k in copy:
        if append_sentence(k):
            edit_sentences_list.append(l.pop(count))
            continue
        count += 1
    return edit_sentences_list

def sentence_card(data : dict):
    temp = [] 
    sentence = data["sentence"]

    temp.append('TARGET DECK: Sentences')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{sentence_to_word_string(sentence)}')
    temp.append('Back: ' + f'{data["translation"]}')
    temp.append(f'Tags: [[{current_name}]] ')
    temp.append('')
    temp.append('END')

    output ='\n'.join(temp)
    write_card(output, f'{sentences_path}\{sentence}.md')
    return

def sentence_card_skipped(sentence : str):
    temp = [] 

    temp.append('TARGET DECK: Sentences')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{sentence_to_word_string(sentence)}')
    temp.append('Back: ')
    temp.append(f'Tags: [[{current_name}]] ')
    temp.append('')
    temp.append('END')

    output ='\n'.join(temp)
    write_card(output, f'{sentences_path}\{sentence}.md')
    return

def word_card(data : dict): 
    temp = []
    word = data['word']
    
    temp.append('TARGET DECK: Words')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{word_to_kanji_string(word)}')
    temp.append('Back: ' + f'{data["definitions"]}')
    temp.append(f'{data["reading"]}')
    temp.append(f'Tags: [[{current_name}]] ')
    temp.append('')
    temp.append('END')

    output ='\n'.join(temp)
    write_card(output, f'{words_path}\{word}.md')
    return
def kanji_card(data : dict):
    temp = []
    kanji = data['kanji_']

    temp.append('TARGET DECK: Kanji')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{kanji}, {data["strokes"]}')
    temp.append('Back: ' + f'{data["keyword"]}')
    temp.append(f'{data["readings"]}')
    temp.append(f'{data["radicals"]}')
    temp.append(f'Tags: [[{current_name}]] ')
    temp.append('')
    temp.append('END')
    
    output ='\n'.join(temp)
    write_card(output, f'{kanji_path}\{kanji}.md')
    return

def kanji_data(kanji : str) -> dict | None:
    request = Kanji.request(kanji)
    output = {}
    try:
        data = request.data
        main_readings = data.main_readings
        output['kanji_'] = kanji
    except:
        return
    try:
        output['keyword'] = data.main_meanings[0]
    except:
        output['keyword'] = ''
    try:
        output['readings'] = *main_readings.kun, *main_readings.on
    except:
        output['readings'] = ''
    try:
        output['strokes'] = data.strokes
    except:
        output['strokes'] = 0
    try:
        output['radicals'] = data.radical.parts
    except:
        output['radicals'] = ''

        return output

    
def word_data(word : str) -> dict:
    request = Word.request(word)
    try:
        defintions = []
        data = request.data[0]
        for defintion in data.senses:
            if defintion.parts_of_speech == ['Wikipedia definition']:
                break
            defintions.append(defintion.english_definitions)
        output = {
            'word' : word,
            'definitions' : str(defintions).replace('[[', '(').replace(']]', ')').replace('[', '(').replace(']', ')').replace("'", ''), # reformatring [['']] as it might become link
            'reading' : data.japanese[0].reading
        }
        return output
    except:
            print(f"Now would be a good time to write to report.txt or something eh? {word} is returning empty")
            return {
            'word' : f'[[{word}]]',
            'definitions' : '()',
            'reading' : ''
        }
def sentence_data(sentence : str) -> dict:
    output = {
        'sentence' : sentence,
        'translation' : gt.translate(sentence, 'en', 'ja'),

    }
    return output
def write_card(lines : str, path : str):
    with open(path, 'w',encoding="utf-8") as file:
        file.write(lines)
    return
def write_sentences_to_content_md(sentences: list[str]):
    this_content_md = content_path + f'/{current_name}.md'
    with open(this_content_md, 'a', encoding='utf8') as file:
        file.write('\n')
        for line in sentences:
            file.write(f'[[{line}]]' + '\n')
        file.close()
    return
def write_sentence_cards(s : str):
    if skip_sentences == False:
        sentence_card(sentence_data(s))
    else:
        sentence_card_skipped(s)
    return

def write_word_cards(word : str):
    word_card(word_data(word))
    return
def write_kanji_cards(kanjis : str):
    kanji_card(kanji_data(kanjis))
    return

def format_link_names (item : str) -> str :
    return f'[[{item}]]\n'
def make_notes ():
    word_punctuation = string.punctuation + r'！＂”“＃＄％＆＇（）＊＋，－．／：；＜＝＞？＠［＼］＾＿｀｛｜｝～、。〃〄々〆〇〈〉《》「」『』【】〒〓〔〕〖〗〘〙〚〛〜〝〞〟〠〡〢〣〤〥〦〧〨〩〪〭〮〯〫〬〰〱〲〳〴〵〶〷〸〹〺〻〼〽〾｟｠｡｢｣､･〿'
    punctuation = ['\n', '.', '?', '!', ' 〪', '。', ' 〭', '！', '．', '？']
    global current_name
    new_content = get_sentences()
    count = 0
    length = len(new_content.items()) - 1
    for name, sentences in new_content.items():
        current_name = name
        temp = ''
        for sentence in sentences:
            temp += sentence
        print(f'sentence lists done for {current_name}')
        words = []
        words_temp = parser(temp)
        
        for w in words_temp:
            if w not in words and w not in word_punctuation and w not in punctuation:
                words.append(w)
        print(f'word lists done for {current_name}')
        kanji_list_temp = [k for k in kanji_set if k in temp]
        kanji_list = []
        for k in kanji_list_temp:
            if k not in kanji_list:
                kanji_list.append(k)
        print(f'kanji lists done for {current_name}')
        print(f'now calling APIs')
        
        with ThreadPoolExecutor() as executor:
            executor.submit(append_content, name)
            future_b = executor.submit(write_to_sentences, sentences)
            future_d = executor.submit(write_to_kanji, kanji_list)
            future_c = executor.submit(write_to_words, words)
            
            
            edit_sentences = future_b.result()
            edit_kanji = future_d.result()
            edit_words = future_c.result()

            kl = list(executor.map(format_link_names, kanji_list))
            wl = list(executor.map(format_link_names, words))
            sl = list(executor.map(format_link_names, sentences))

            add_new_stuff(kl, wl, sl)
            if len(edit_kanji) > 0:
                executor.map(edit_kanji_tags, edit_kanji)             
            if len(edit_sentences) > 0:
                executor.map(edit_sentence_tags, edit_sentences) 
            if len(edit_words) > 0:
                executor.map(edit_words_tags, edit_words) 

            executor.map(write_kanji_cards, kanji_list)
            executor.submit(write_sentences_to_content_md, sentences)
            executor.map(write_sentence_cards, sentences)
            executor.map(write_word_cards, words)

            executor.shutdown(wait=True)
        if count < length:
            print("API calls are done, next loop")
            continue
        print("It is done, Enjoy your notes")
    
class Flashcard:
    def __init__ (self, front : str, back : str):
        self.Front = front.replace('[', '').replace(']', '')
        self.Back = back
        self.Cloze = ''
        if '[[' in front and len(front.replace('[[', '').replace(']', '')) > 1:
            temp = front.replace('[[', '{').replace(']]', '}}')
            result = ''
            count = 1
            for t in temp:
                if t != '{':
                    result += t
                    continue
                result += '{{' + f'c{count}::'
                count += 1
            self.Cloze = result
def files_to_flashcard_class(file_paths : list[int]) -> list[dict]:
    output = []
    lines = []
    for path in file_paths:
        with open(path, 'r', encoding="utf-8") as file:
            lines = file.readlines()
        front_index = 3 # need a better way of setting front_index

        if front_index < 0: continue
        back_index, tag_index = 0, 0
        for i in range(front_index, len(lines)):
            if 'Back:' not in lines[i]: continue
            lines[i].replace("Back: ", "")
            back_index = i
            break
        if back_index == 0: continue

        for i in range(len(lines)-1, back_index, -1):
            if 'Tags: [[' not in lines[i]: continue
            tag_index = i
            break
        front_range = range(front_index, back_index)
        back_range = range(back_index, tag_index)
        front, back = '', ''
        for i in front_range:
            front += lines[i]
        for i in back_range:
            
            back += lines[i]
        output.append(Flashcard(front, back).__dict__)
            
    return output
def flashcards_to_csv (flashcards : list[dict], csv_file_path : str, cloze_path : str):
    regular_fieldnames = ['Front', 'Back']
    cloze_fieldnames = ['Cloze', 'Back']
    with open(csv_file_path, 'w', encoding="utf-8", newline='') as regular_csv_file, \
        open(cloze_path, 'w', encoding="utf-8", newline='') as cloze_csv_file:
        regular_csv_writer = csv.DictWriter(regular_csv_file, regular_fieldnames)
        cloze_csv_writer = csv.DictWriter(cloze_csv_file, cloze_fieldnames)
        for card in flashcards:
            back = card['Back']
            cloze_csv_writer.writerow({'Cloze': card['Cloze'], 'Back': back})
            regular_csv_writer.writerow({'Front': card['Front'], 'Back': back})
def make_csvs ():
    input_csv_sentences = glob.glob(f"{sentences_path}\*.md")
    input_csv_words = glob.glob(f"{words_path}\*.md")
    input_csv_kanji = glob.glob(f"{kanji_path}\*.md")
    flashcards_to_csv(files_to_flashcard_class(input_csv_sentences), f'{csv_path}\Sentences.csv', f'{csv_path}\Sentences_cloze.csv')
    flashcards_to_csv(files_to_flashcard_class(input_csv_words), f'{csv_path}\Words.csv', f'{csv_path}\Words_cloze.csv')
    flashcards_to_csv(files_to_flashcard_class(input_csv_kanji), f'{csv_path}\Kanji.csv', f'{csv_path}\Kanji_cloze.csv')
def ask_for_translations(n : int):
    global skip_sentences
    print(r"This program uses Google Translate for sentence translations. If you have more accurate translations, you might want to skip this step.")
    print(r"Include sentence translations?")
    print(r"WARNING YOU WILL HAVE TO QUIT IF YOU ENTER THIS INCORRECTLY.")
    answer = input(r'Y (for yes) or N (for no)').lower()
    match answer:
        case 'y':
            print(r"Alright, translations will be there.")
            ready(n)
        case 'n':
            print(r"Alright, no translations")
            skip_sentences = True
            ready(n)
        case _:
            print("that's not an answer I understand.")
            ask_for_translations()
def ask_for_csvs ():
    print(r"do you want the .csv files (for Anki)?")
    answer = input(r'Y (for yes) or N (for no)').lower()
    match answer:
        case 'y':
            just_csvs()
        case 'n':
            ask_for_translations(0)
        case _:
            print("that's not an answer I understand.")
            ask_for_csvs()
def just_csvs():
    print(r"do you only want to make the .csv files (No new/overwriting markdown notes)?")
    answer = input(r'Y (for yes) or N (for no)').lower()
    match answer:
        case 'y':
            ready(2)
        case 'n':
            ask_for_translations(1)
        case _:
            print("that's not an answer I understand.")
            just_csvs()
def ready(n : int):
    match n:
        case 0:
            print("So you want only the NOTES. Great :^)")
        case 1:
            print("so you want both the NOTES and the CSV files for Anki? Wonderful :-^)")
        case 2:
            print("so you want only the CSV files for Anki? Beautiful :^D")
        case _:
            print("We did something we shouldn't have, maybe it's a bitflip, maybe it's maybeline, sorry :C")
            ask_for_csvs()  
    answer = input(r'Y (for yes) or N (for no)').lower()
    match answer:
        case 'y':
            match n:
                case 0:
                    print("Creating just notes")
                    make_notes()
                case 1:
                    print("Creating notes and csv files")
                    make_notes()
                    make_csvs()
                case 2:
                    print("creating csv files")
                    make_csvs()
                case _:
                    print("We did something we shouldn't have, maybe it's a bitflip, maybe it's fitblip, sorry :C")
                    ask_for_csvs()  
        case 'n':
            print(r" Mistakes happen, we'll start from the begining again. <(^~^)>")
            ask_for_csvs()
        case _:
            print("that's not an answer I understand.")
            ready(n)


def main():
    ask_for_csvs()
if __name__ == '__main__':
    main()