tag = "ikimono-gakari_sakura"
test = "TARGET DECK: Kanji\nSTART\nBasic\n中, 4\nBack: in\n('なか', 'うち', 'あた.る', 'チュウ')\n['｜', '口']\nTags: [[Again_Yui]] \n\nEND"
lines = test.split('\n')
for line in lines:
    if 'Tags: ' in line:
        lines[lines.index(line)] = line[:-2] + '] ' + f'[[{tag}]] '

print(*lines)