// Package ref defines a reference implementation of Tribbler service.
package triblab

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
	"trib"
        "trib/colon"
        "strconv"
        "strings"
)

type Server struct {
	lock  sync.Mutex
        BS trib.BinStorage
}

var _ trib.Server = new(Server)

func NewServer() *Server {
	ret := &Server{
	}
	return ret
}

func (self *Server) findUser(user string) bool {

        nameL, _ := self.listUser()

        for _, name := range nameL {
          if name == user {
            return true
          }
        }
        return false
}

func (self *Server) listUser() ([]string,error) {
       //fmt.Println("come to listUser")
        var keyList = new(trib.List)
        storage := self.BS.Bin("SERVER")
        pattern := trib.Pattern{Prefix: "", Suffix: "USERS"}
        storage.Keys(&pattern, keyList)
        ret := make([]string, 0, len(keyList.L))
        for _, keys := range keyList.L {
          var user string
          storage.Get(keys, &user)
          ret = append(ret, user)
        }

	sort.Strings(ret)
	if len(ret) > trib.MinListUser {
		ret = ret[:trib.MinListUser]
	}

	return ret, nil
}

func (self *Server) following(who string) ([]string, error) {
       //fmt.Println("come to following")
	found := self.findUser(who)
	if !found {
	  return nil, fmt.Errorf("user %q does not exist", who)
	}

       storage := self.BS.Bin("SERVER")
       var keyList = new(trib.List)
       pattern := trib.Pattern{Prefix: who, Suffix: "Following"}
       storage.Keys(&pattern, keyList)
       ret := make([]string, 0, len(keyList.L))
       for _, key := range keyList.L {
         var name string
         storage.Get(key, &name)
         ret = append(ret,name)
       }

	return ret, nil
}


func (self *Server) isFollowing(who, whom string) (bool, error) {
       //fmt.Println("come to isFollowing")
	if who == whom {
		return false, fmt.Errorf("checking the same user")
	}

        found := self.findUser(who)
	if !found {
		return false, fmt.Errorf("user %q does not exist", who)
	}
        found = self.findUser(whom)
	if !found {
		return false, fmt.Errorf("user %q does not exist", whom)
	}

       storage := self.BS.Bin("SERVER")
       var keyList = new(trib.List)
       pattern := trib.Pattern{Prefix: who, Suffix: "Following"}
       storage.Keys(&pattern, keyList)

       for _, key := range keyList.L {
         var name string
         storage.Get(key, &name)
         if(name == whom) {
           return true, nil
         }
       }

       return false, nil
}

func (self *Server) getClock(user string, clock uint64) uint64 {
       //fmt.Println("come to getClock")
  storage := self.BS.Bin(user)
  var ret uint64
  storage.Clock(clock,&ret)
  return ret
}

func (self *Server) SignUp(user string) error {
        //fmt.Println("come to SignUp()")
	if len(user) > trib.MaxUsernameLen {
		return fmt.Errorf("username %q too long", user)
	}

	if !trib.IsValidUsername(user) {
		return fmt.Errorf("invalid username %q", user)
	}

	self.lock.Lock()
	defer self.lock.Unlock()

        found := self.findUser(user)
        if found {
	  return fmt.Errorf("user %q already exists", user)
        }

        kv := trib.KeyValue{Key: user+"USERS", Value: user}
        storage := self.BS.Bin("SERVER")
        var succ bool
        storage.Set(&kv, &succ)
        if !succ {
          return fmt.Errorf("Add user %q failed", user)
        }
	return nil
}

func (self *Server) ListUsers() ([]string, error) {
        //fmt.Println("come to ListUsers()")
	self.lock.Lock()
	defer self.lock.Unlock()

        var keyList = new(trib.List)
        storage := self.BS.Bin("SERVER")
        pattern := trib.Pattern{Prefix: "", Suffix: "USERS"}
        storage.Keys(&pattern, keyList)
        ret := make([]string, 0, len(keyList.L))
        for _, keys := range keyList.L {
          var user string
          storage.Get(keys, &user)
          ret = append(ret, user)
        }

	sort.Strings(ret)
	if len(ret) > trib.MinListUser {
		ret = ret[:trib.MinListUser]
	}

	return ret, nil
}

func (self *Server) IsFollowing(who, whom string) (bool, error) {
        //fmt.Println("come to IsFollowing()")
	if who == whom {
		return false, fmt.Errorf("checking the same user")
	}

	self.lock.Lock()
	defer self.lock.Unlock()

        found := self.findUser(who)
	if !found {
		return false, fmt.Errorf("user %q does not exist", who)
	}
        found = self.findUser(whom)
	if !found {
		return false, fmt.Errorf("user %q does not exist", whom)
	}

       storage := self.BS.Bin("SERVER")
       var keyList = new(trib.List)
       pattern := trib.Pattern{Prefix: who, Suffix: "Following"}
       storage.Keys(&pattern, keyList)

       for _, key := range keyList.L {
         var name string
         storage.Get(key, &name)
         if(name == whom) {
           return true, nil
         }
       }

       return false, nil
}

func (self *Server) Follow(who, whom string) error {
       //fmt.Println("come to Follow()")
       if who == whom {
         return fmt.Errorf("checking the same user")
       }

       self.lock.Lock()
       defer self.lock.Unlock()

       found := self.findUser(who)
       if !found {
         return fmt.Errorf("user %q does not exist", who)
       }
       found = self.findUser(whom)
       if !found {
         return fmt.Errorf("user %q does not exist", whom)
       }

       storage := self.BS.Bin("SERVER")
       var keyList = new(trib.List)
       pattern := trib.Pattern{Prefix: who, Suffix: "Following"}
       storage.Keys(&pattern, keyList)

       count := 0
       for _, key := range keyList.L {
         var name string
         storage.Get(key, &name)
         if name == whom {
           return fmt.Errorf("user %q already following %q", who, whom)
         }
           count = count + 1
       }

       if count >= trib.MaxFollowing {
		return fmt.Errorf("user %q is following too many users")
       }

       kv := trib.KeyValue{Key:who+whom+"Following", Value: whom}
       var suc bool
       storage.Set(&kv, &suc)
       if !suc {
               return fmt.Errorf("user %q failed to follow %q", who, whom)
       }
	return nil
}

func (self *Server) Unfollow(who, whom string) error {
       //fmt.Println("come to UnFollow()")
       found, err := self.isFollowing(who, whom)

       if err != nil {
         return err
       }

       if !found {
	return fmt.Errorf("user %q is not following %q", who, whom)
       }

       storage := self.BS.Bin("SERVER")
       kv := trib.KeyValue{Key:who+whom+"Following", Value: ""}
       var suc bool
       storage.Set(&kv, &suc)
       if !suc {
         return fmt.Errorf("user %q failed to unfollow %q", who, whom)
       }
       return nil
}

func (self *Server) Following(who string) ([]string, error) {
       //fmt.Println("come to Following()")
	self.lock.Lock()
	defer self.lock.Unlock()

	found := self.findUser(who)
	if !found {
	  return nil, fmt.Errorf("user %q does not exist", who)
	}

       storage := self.BS.Bin("SERVER")
       var keyList = new(trib.List)
       pattern := trib.Pattern{Prefix:who, Suffix: "Following"}
       storage.Keys(&pattern, keyList)
       ret := make([]string, 0, len(keyList.L))
       for _, key := range keyList.L {
         var name string
         storage.Get(key, &name)
         ret = append(ret,name)
       }

	return ret, nil
}

func (self *Server) Post(user, post string, c uint64) error {
       //fmt.Println("come to Post()")
	if len(post) > trib.MaxTribLen {
		return fmt.Errorf("trib too long")
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	found := self.findUser(user)
	if !found {
		return fmt.Errorf("user %q does not exist", user)
	}

        seq := self.getClock(user, c)
	if seq == math.MaxUint64 {
		panic("run out of seq number")
	}
        storage := self.BS.Bin(user)
        newPost := colon.Escape(post)
        kv := trib.KeyValue {Key: "Posts", Value: strconv.FormatUint(seq, 10)+"|1"+time.Now().Format(time.RFC3339)+"|2"+newPost}
        var succ bool
        storage.ListAppend(&kv, &succ)
        if !succ {
          return fmt.Errorf("user %q post: %q failed", user, post)
        }
	return nil
}

func (self *Server) Home(user string) ([]*trib.Trib, error) {
       //fmt.Println("come to Home()")
	self.lock.Lock()
	defer self.lock.Unlock()

        following, err := self.following(user)
        if err != nil {
          return nil, err
        }

        home :=  make([]*trib.Trib, 0, 4096)
        var addSelfPostList []string
        addSelfPostList = append(following, user)
        for _, name := range addSelfPostList {
          storage := self.BS.Bin(name)
          var postList = new(trib.List)
          storage.ListGet("Posts", postList)
          for _, post := range postList.L {
            s := strings.Split(post, "|1")
            clock := s[0]
            s1 := s[1]
            timeAndPost := strings.Split(s1, "|2")
            times := timeAndPost[0]
            post := colon.Unescape(timeAndPost[1])
            parseTime, _ := time.Parse(time.RFC3339,times)
            parseClock, _ := strconv.ParseUint(clock, 10, 64)
            t := &trib.Trib{User: name, Message: post, Time: parseTime, Clock: parseClock}
            home = append(home, t)
          }
        }
        ntrib := len(home)
        start := 0
        sortTrib(home)
        if ntrib > trib.MaxTribFetch {
          start = ntrib-trib.MaxTribFetch
        }
	return home[start:], nil
}

func older (a,b *trib.Trib) bool {
  if a.Clock < b.Clock {
    return true
  }
  if a.Clock > b.Clock {
    return false
  }
  if a.Time.Before(b.Time) {
    return true
  }
  if !a.Time.Before(b.Time) {
    return false
  }

  userA, _ := strconv.Atoi(a.User)
  userB, _ := strconv.Atoi(b.User)
  if userA < userB {
    return true
  }
  if userA > userB {
    return false
  }

  messageA, _ := strconv.Atoi(a.Message)
  messageB, _ := strconv.Atoi(b.Message)
  if messageA < messageB {
    return true
  }
  return false
}

func sortTrib (t []*trib.Trib) {
  //fmt.Println("get in sortTrib()")
  /*fmt.Println("-------before---------")
  for _, eachTrib1 := range t {
    fmt.Println(eachTrib1.Message)
    fmt.Println(eachTrib1.Clock)
  }*/
  for i, eachTrib1 := range t {
    oldestTrib := eachTrib1
    index := i
    for j, eachTrib2 := range t[i+1:] {
      if !older(oldestTrib, eachTrib2) {
        oldestTrib = eachTrib2
        index = j
      }
    }
    temp := &trib.Trib{User: eachTrib1.User, Message: eachTrib1.Message, Clock: eachTrib1.Clock, Time: eachTrib1.Time}
    eachTrib1.User = oldestTrib.User
    eachTrib1.Message = oldestTrib.Message
    eachTrib1.Clock = oldestTrib.Clock
    eachTrib1.Time = oldestTrib.Time
    t[index].User = temp.User
    t[index].Message = temp.Message
    t[index].Clock = temp.Clock
    t[index].Time = temp.Time
  }
  /*
  fmt.Println("-------after---------")
  for _, eachTrib1 := range t {
    fmt.Println(eachTrib1.Message)
    fmt.Println(eachTrib1.Clock)
  }*/
}

func (self *Server) Tribs(user string) ([]*trib.Trib, error) {
    //   fmt.Println("come to Tribs()")
	self.lock.Lock()
	defer self.lock.Unlock()

	found := self.findUser(user)
	if !found {
		return nil, fmt.Errorf("user %q does not exist", user)
	}

        tribs := make([]*trib.Trib, 0, 4096)
        var postList = new(trib.List)
        storage := self.BS.Bin(user)
        storage.ListGet("Posts", postList)
        for _, post := range postList.L {
            s := strings.Split(post, "|1")
            clock := s[0]
            s1 := s[1]
            timeAndPost := strings.Split(s1, "|2")
            times := timeAndPost[0]
            post := colon.Unescape(timeAndPost[1])
            parseTime, _ := time.Parse(time.RFC3339,times)
            parseClock, _ := strconv.ParseUint(clock, 10, 64)
            t := &trib.Trib{User: user, Message: post, Time: parseTime, Clock: parseClock}
            tribs = append(tribs, t)
        }
        ntrib := len(tribs)
        start := 0
        sortTrib(tribs)
        if ntrib > trib.MaxTribFetch {
          start = ntrib-trib.MaxTribFetch
        }
	return tribs[start:], nil
}
