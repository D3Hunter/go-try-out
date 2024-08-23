package main

import (
	"container/heap"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const MaxValue = 3
const ValueCnt = 4

type node struct {
	values []int
}

func (n *node) String() string {
	var sb strings.Builder
	for _, v := range n.values {
		sb.WriteString(strconv.Itoa(v))
	}
	return sb.String()
}

type distNode struct {
	dist int
	*node
}

type distNodeHeap struct {
	nodes []*distNode
	index map[string]int
}

func newNodeHeap() *distNodeHeap {
	return &distNodeHeap{
		nodes: make([]*distNode, 0),
		index: make(map[string]int),
	}
}

func (n *distNodeHeap) FixNode(nd *distNode) {
	if idx, ok := n.index[nd.String()]; ok {
		heap.Fix(n, idx)
	}
}

func (n *distNodeHeap) Len() int {
	return len(n.nodes)
}

func (n *distNodeHeap) Less(i, j int) bool {
	return n.nodes[i].dist < n.nodes[j].dist
}

func (n *distNodeHeap) Swap(i, j int) {
	n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
	n.index[n.nodes[i].String()] = i
	n.index[n.nodes[j].String()] = j
}

func (n *distNodeHeap) Push(x interface{}) {
	nd := x.(*distNode)
	n.nodes = append(n.nodes, nd)
	n.index[nd.String()] = len(n.nodes) - 1
}

func (n *distNodeHeap) Pop() interface{} {
	item := n.nodes[len(n.nodes)-1]
	n.nodes = n.nodes[:len(n.nodes)-1]
	delete(n.index, item.String())
	return item
}

func (n distNode) String() string {
	return fmt.Sprintf("%d%d%d%d", n.values[0], n.values[1], n.values[2], n.values[3])
}

var TargetNode = &node{[]int{MaxValue, MaxValue, MaxValue, MaxValue}}

var source = &node{[]int{1, 2, 3, 1}}

func genGraph() (map[string]*node, map[string][]*node) {
	edges := make(map[string][]*node)
	nodes := make(map[string]*node)
	nodes[source.String()] = source
	nodes[TargetNode.String()] = TargetNode

	var grayNodes []*node
	grayNodes = append(grayNodes, source)
	grayNodeMap := make(map[string]bool)
	grayNodeMap[source.String()] = true
	visitedMap := make(map[string]bool)
	for len(grayNodes) > 0 {
		item := grayNodes[0]
		grayNodes = grayNodes[1:]
		delete(grayNodeMap, item.String())
		visitedMap[item.String()] = true
		if item.String() == TargetNode.String() {
			break
		}

		for i := 0; i < ValueCnt; i++ {
			next := &node{make([]int, ValueCnt)}
			copy(next.values, item.values)
			next.values[i] = next.values[i]%MaxValue + 1
			if i == 0 {
				next.values[i+1] = next.values[i+1]%MaxValue + 1
			} else if i == ValueCnt-1 {
				next.values[i-1] = next.values[i-1]%MaxValue + 1
			} else {
				next.values[i-1] = next.values[i-1]%MaxValue + 1
				next.values[i+1] = next.values[i+1]%MaxValue + 1
			}
			if n, ok := nodes[next.String()]; ok {
				next = n
			} else {
				nodes[next.String()] = next
			}
			edges[item.String()] = append(edges[item.String()], next)
			if !visitedMap[next.String()] && !grayNodeMap[next.String()] {
				grayNodes = append(grayNodes, next)
				grayNodeMap[next.String()] = true
			}
		}
	}
	return nodes, edges
}

func main() {
	nodes, edges := genGraph()
	prevPath := make(map[string]*distNode)
	minHeap := newNodeHeap()
	heap.Init(minHeap)
	distNodeMap := make(map[string]*distNode)
	for _, n := range nodes {
		dist := math.MaxInt32
		if n == source {
			dist = 0
		}
		distN := &distNode{dist: dist, node: n}
		heap.Push(minHeap, distN)
		distNodeMap[distN.String()] = distN
	}

	for minHeap.Len() > 0 {
		n := heap.Pop(minHeap).(*distNode)
		if neighbors, ok := edges[n.String()]; ok {
			for _, neighbor := range neighbors {
				alt := n.dist + 1
				distN := distNodeMap[neighbor.String()]
				if alt < distN.dist {
					distN.dist = alt
					prevPath[distN.String()] = n
					minHeap.FixNode(distN)
				}
			}
		}
	}
	u := &distNode{node: TargetNode}
	var reversePath []string
	if prevPath[u.String()] != nil || u.String() == source.String() {
		for u != nil {
			reversePath = append(reversePath, u.String())
			u = prevPath[u.String()]
		}
	}

	for i := len(reversePath) - 1; i >= 0; i-- {
		fmt.Printf("%s ", reversePath[i])
	}
	fmt.Println()
}
