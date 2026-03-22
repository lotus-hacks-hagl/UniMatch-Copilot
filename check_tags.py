import re
import os

file_path = r'd:\CODE\UniMatch-Copilot\frontend\src\views\CaseDetailView.vue'
if not os.path.exists(file_path):
    print(f"File not found: {file_path}")
    exit(1)

with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# Simple tag balancer
# Regex to find <tag or </tag
# We ignore some standard self-closing tags and assume others are paired
tags = re.findall(r'<(/?[\w-]+)', content)
stack = []
ignored_self_closing = {'img', 'br', 'hr', 'input', 'link', 'meta', 'path', 'circle', 'rect', 'svg'}

for tag in tags:
    if tag.startswith('/'):
        closing = tag[1:]
        if not stack:
            print(f"Excess closing tag: </{closing}>")
        else:
            opening = stack.pop()
            if opening.lower() != closing.lower():
                print(f"Mismatched tag: <{opening}> closed by </{closing}>")
    else:
        # Ignore self-closing tags (rough)
        if tag.lower() not in ignored_self_closing:
            stack.append(tag)

if stack:
    # Filter out common Vue components that might be self-closing in template but the script doesn't know
    # Actually, Transition and TransitionGroup should be closed.
    print(f"Unclosed tags: {stack}")
else:
    print("Tags are balanced (roughly)")
