import collections
import html
import logging
import sys

import ruamel.yaml

# This script parses out the comments from a values.yaml file, generating
#   a Markdown table of key to comment for documentation purposes.
# To run this:
#  - install python3
#  - pip3 install ruamel.yaml
#  - python3 create_table_from_values_comments.py values.yaml

class Node:
    def __init__(self, path):
        self.path = path
        self.comment = None
        self.dict = {}
        self.value = None

    def get_path(self):
        return '.'.join(self.path)

    def get_type(self):
        if len(self.dict) > 0 and self.value is not None:
            raise ValueError("node both has dict and value")
        if len(self.dict) > 0:
            return "dict"
        if self.value is not None:
            return "leaf"
        # TODO array
        return "unknown"

    def set_value(self, value):
        if len(self.dict) > 0:
            raise ValueError("can not set value: non-empty dict")
        if self.value is not None:
            raise ValueError("can not set value: already set ({}, {}, {})".format(self.path, self.value, value))
        self.value = value

    def set_comment(self, comment):
        logging.debug("setting comment for %s to %s", self.path, comment)
        if self.comment is not None:
            raise ValueError("can not set comment: already set ({}, {}, {})".format(self.path, self.comment, comment))
        self.comment = comment

    def add_child(self, key):
        logging.debug("add child to %s, %s", self.path, key)
        if self.value is not None:
            raise ValueError("can not add child: value already set")
        if key not in self.dict:
            self.dict[key] = Node(self.path + [key])
        return self.dict[key]

    def add_path(self, path):
        node = self
        for key in path:
            node = node.add_child(key)
        return node


def print_tree(node):
    path = node.path
    indent = '  ' * len(path)
    node_type = node.get_type()
    if node_type == "leaf":
        print(indent + " " + str(path) + " " + str(node.value))
    elif node_type == "dict":
        print(indent + " " + str(path))
        for (k, v) in node.dict.items():
            print_tree(v)
    elif node_type == "unknown":
        print(indent + " UNKNOWN " + str(path))
    else:
        raise ValueError("invalid type: {}".format(node_type))

def clean_comment(comment):
    if comment is None or comment.strip(' \n') == '':
        return ''
    return html.escape(comment).replace('\\', '&#92;').replace('\n', '<br />')

def clean_value(value):
    if value is None:
        return ''
    elif value == '':
        return '`""`'
    elif value is True:
        return '`"true"`'
    elif value is False:
        return '`"false"`'
    elif isinstance(value, str):
        return '`"{}"`'.format(value)
    return '`{}`'.format(str(value))

def clean_line(line):
    cleaned = '\n'.join(p.strip(' \n').strip('#').rstrip('\n') for p in line.split('\n'))
    logging.debug("cleaned? <%s>, %s", cleaned, len(cleaned))
    return cleaned

def get_token(thing):
    if isinstance(thing, list):
        return ' '.join(clean_line(l.value) for l in thing).strip('\n')
    else:
        return clean_line(thing.value).strip('\n')

def clean_comment_nodes(lt):
    return [get_token(t) if t is not None else t for t in lt]

def print_table(node):
    path = node.get_path()
    node_type = node.get_type()
    if node_type == "leaf":
        print("| {} | {} | <pre>{}</pre> |".format(path, clean_value(node.value), clean_comment(node.comment)))
    elif node_type == "dict":
        print("| {} | N/A | <pre>{}</pre> |".format(path, clean_comment(node.comment)))
        for (k, v) in sorted(node.dict.items(), key=lambda x: x[0]):
            print_table(v)
    elif node_type == "unknown":
        print("| {} | | <pre>{}</pre> |".format(path, clean_comment(node.comment)))
    else:
        raise ValueError("invalid type: {}".format(node_type))

def match_comments_to_keypath(d, stack, comments, node):
    logging.debug(("push stack state", stack, list(comments)))
    if len(stack) > 0:
        if len(comments) == 0:
            logging.debug(("no comment found", stack, comments))
        else:
            next_comment = comments.popleft()
            node.set_comment(next_comment)

            logging.debug(("found comment", stack, next_comment))

    # dictionaries
    if isinstance(d, dict):

        if d.ca.comment is not None:
            logging.debug(("dictionary comment", stack, d.ca.comment))

        # iterate through contents of container
        for key, val in d.items():
            new_stack = stack + (key,)

            logging.debug('  ' * len(stack) + "found dictionary child: %s", new_stack)

            # save comment for this key/val, if it's present
            w = x = y = z = None
            if key in d.ca.items:
                w, x, y, z = clean_comment_nodes(d.ca.items[key])
                logging.debug(("look at my comment", new_stack, d.ca.items[key], w, x, y, z))

            if w is not None:
                raise ValueError("no idea what this means")
            if x is not None:
                comments.append(x)
                logging.debug(("enqueue comment", new_stack, x))
            if z is not None:
                comments.append(z)
                logging.debug(("enqueue comment", new_stack, z))

            # here we're hitting the *key* for a *leaf*
            #   so we go back and grab the last comment
            current_node = node.add_path([key])

            # recur if this is a container
            match_comments_to_keypath(val, new_stack, comments, current_node)

            if y is not None:
                comments.append(y)
                logging.debug(("enqueue comment", new_stack, y))
                logging.debug(('before recurse', new_stack, d.ca.items[key]))
    # lists
    elif isinstance(d, list):
        logging.debug('  ' * len(stack) + "found list: %s", stack)
        # save comment on list, if it's present
        if d.ca.comment is not None:
            logging.debug('  ' * len(stack) + "comment from list container: %s", d.ca.comment)

        # iterate through list elements
        for idx, item in enumerate(d):

            raise Exception("TODO -- implement this case")
    # primitives
    else:
        logging.debug(("value", stack, d))
        # current_node = node.add_path(stack)
        node.set_value(d)

    logging.debug(("pop", stack))


def run():
    path = sys.argv[1]
    log_level = logging.INFO
    if len(sys.argv) == 3:
        log_level = sys.argv[2]
    elif len(sys.argv) > 3:
        raise ValueError("expected 1 or 2 command line args, found {}".format(len(sys.argv) - 1))
    logging.basicConfig(encoding='utf-8', level=log_level)

    with open(path) as f:
        yaml_str = f.read()

    data = ruamel.yaml.YAML().load(yaml_str)

    node = Node([])
    logging.debug(data)

    comments = collections.deque()
    if data.ca.comment is not None:
        comments.append(get_token(data.ca.comment[1]))
    match_comments_to_keypath(data, (), comments, node)

    # print_tree(node)

    print("# Root values")
    print("| Key | Default value | Description |")
    print("|----|----|----|")
    for (k, v) in sorted(node.dict.items(), key=lambda x: x[0]):
        node_type = v.get_type()
        if node_type == "leaf":
            print("| {} | {} | <pre>{}</pre> |".format(v.get_path(), clean_value(v.value), clean_comment(v.comment)))
        elif node_type == "unknown":
            print("| {} | | <pre>{}</pre> |".format(v.get_path(), clean_comment(v.comment)))

    for (k, v) in sorted(node.dict.items(), key=lambda x: x[0]):
        node_type = v.get_type()
        if node_type == "dict":
            print("\n# {} values".format(v.get_path()))
            print("| Key | Default value | Description |")
            print("|----|----|----|")
            print_table(v)

run()
