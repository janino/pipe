// Pipe - A small and beautiful blogging platform written in golang.
// Copyright (C) 2017, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/b3log/pipe/model"
	"github.com/b3log/pipe/util"
	"github.com/ghodss/yaml"
)

type MarkdownFile struct {
	Filename string
	Filepath string
	Content  string
}

func ImportMarkdowns(mdFiles []*MarkdownFile, authorID, blogID uint) {
	succCnt, failCnt := 0, 0
	fails := []string{}
	for _, mdFile := range mdFiles {
		article := parseArticle(mdFile)
		article.AuthorID = authorID
		article.BlogID = blogID

		if err := Article.AddArticle(article); nil != err {
			failCnt++
			fails = append(fails, mdFile.Filename)
			logger.Errorf("import article failed: " + err.Error())

			continue
		}

		os.Rename(mdFile.Filepath, mdFile.Filepath+"."+strconv.Itoa(int(article.ID)))
		succCnt++
	}

	if 0 == succCnt && 0 == failCnt {
		return
	}

	logBuilder := "[" + strconv.Itoa(succCnt) + "] imported, [" + strconv.Itoa(failCnt) + "] failed"
	if 0 < failCnt {
		logBuilder += ": \n"
		for _, fail := range fails {
			logBuilder += "    " + fail + "\n"
		}
	} else {
		logBuilder += " :p"
	}

	logger.Info(logBuilder)
}

func parseArticle(mdFile *MarkdownFile) *model.Article {
	util.Recover()

	content := strings.TrimSpace(mdFile.Content)
	frontMatter := strings.Split(content, "---")[0]
	if "" == frontMatter {
		content = strings.Split(content, "---")[1]
		frontMatter = strings.Split(content, "---")[0]
	}

	ret := &model.Article{}

	m := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(frontMatter), &m)
	if nil != err {
		ext := filepath.Ext(mdFile.Filename)
		ret.Title = strings.Split(mdFile.Filename, ext)[0]
		ret.Content = content
		ret.Commentable = true
		ret.Tags = "笔记"

		return ret
	}

	ext := filepath.Ext(mdFile.Filename)
	title := strings.Split(mdFile.Filename, ext)[0]
	if t, ok := m["title"]; ok {
		title = strings.TrimSpace(t.(string))
	}
	ret.Title = title

	content = strings.TrimSpace(strings.Split(content, frontMatter)[1])
	if strings.HasPrefix(content, "---") {
		content = strings.Split(content, "---")[1]
		content = strings.TrimSpace(content)
	}
	ret.Content = content

	permalink := ""
	if p, ok := m["permalink"]; ok {
		permalink = strings.TrimSpace(p.(string))
	}
	ret.Path = permalink

	tags := parseTags(&m)
	ret.Tags = tags
	ret.Commentable = true

	return ret
}

func parseTags(m *map[string]interface{}) string {
	frontMatter := *m
	tags := frontMatter["tags"]
	if nil == tags {
		tags = frontMatter["category"]
	}
	if nil == tags {
		tags = frontMatter["categories"]
	}
	if nil == tags {
		tags = frontMatter["keyword"]
	}
	if nil == tags {
		tags = frontMatter["keywords"]
	}
	if nil == tags {
		return "笔记"
	}

	ts := tags.([]interface{})
	tagStrs := []string{}
	for _, t := range ts {
		tagStrs = append(tagStrs, t.(string))
	}

	return strings.Join(tagStrs, ",")
}
