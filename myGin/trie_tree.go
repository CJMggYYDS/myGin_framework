package myGin

import (
	"fmt"
	"strings"
)

/*
	** 动态路由匹配 **
	选择使用前缀树结构(Trie tree)来实现动态路由匹配
	匹配策略：
	* 参数匹配':' 例如 /p/:lang/doc，可以匹配 /p/c/doc 和 /p/go/doc ;
	* 通配'*' 例如 /static/*filepath，可以匹配/static/fav.ico，也可以匹配/static/js/jQuery.js，
			这种模式常用于静态服务器，能够递归地匹配子路径。

*/

// 定义trie树节点结构
type node struct {
	pattern  string  //是否是一个完整的路由，如果不是则为""
	part     string  //url块值,路由中用/分割的部分
	children []*node //子节点
	isWild   bool    //是否模糊匹配,part中含有:或者*时为true
}

// 为节点定义一个toString函数
func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t", n.pattern, n.part, n.isWild)
}

// 节点插入操作(递归实现)
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// 搜索查找操作(递归实现)
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

// 遍历,获得所有完整的url节点
func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

// 查找第一个根据part匹配成功的节点, 用于插入操作
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 查找所有匹配成功的节点, 用于查找操作
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
