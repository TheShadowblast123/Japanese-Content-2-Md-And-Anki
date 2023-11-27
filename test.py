kanji_range_1, kanji_range_2, kanji_range_3 = (0x3400, 0x4DB5),(0x4E00,0x9FCB), (0xF900, 0xFA6A)
kanji_set_1, kanji_set_2, kanji_set_3 = [chr(c) for c in range(*kanji_range_1)], [chr(c) for c in range(*kanji_range_2)], [chr(c) for c in range(*kanji_range_3)]
kanji_set = kanji_set_1 + kanji_set_2 + kanji_set_3
import re

test = '飛び立つ'
title = '[[飛]] び [[立]] つ'
def check_title(title, test):
    return test == title.replace(' ', '').replace('[', '').replace(']', '')
print(check_title(title, test))