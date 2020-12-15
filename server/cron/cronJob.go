package cron

import (
	goCron "github.com/robfig/cron/v3"
)

type MyJob struct {
	Id int64
	EntryID goCron.EntryID
	Job *goCron.Cron
}

type Callback func(id int64,job *MyJob)

var jobCache []MyJob


func Remove(id int64)  {
	for i,job := range jobCache {
		if job.Id == id {
			job.Job.Stop()
			job.Job.Remove(job.EntryID)
			jobCache = append(jobCache[:i], jobCache[i+1:]...)
			break
		}
	}
}

func Stop(id int64)  {
	for _,job := range jobCache {
		if job.Id == id {
			job.Job.Stop()
			job.Job.Remove(job.EntryID)
			break
		}
	}
}

func Hold(id int64)  {
	for _,job := range jobCache {
		if job.Id == id {
			job.Job.Stop()
			break
		}
	}
}

func Run(id int64,callback Callback)  {
	hasJob := false
	for _,item := range jobCache {
		if item.Id == id {
			item.Job.Stop()
			if callback != nil {
				item.Job.Remove(item.EntryID)
				callback(id,&item)
			}else {
				item.Job.Start()
			}
			hasJob = true
			break
		}
	}
	if !hasJob {
		job := goCron.New(goCron.WithParser(goCron.NewParser(
			goCron.SecondOptional | goCron.Minute | goCron.Hour | goCron.Dom | goCron.Month | goCron.Dow,
		)))
		myJob := MyJob{
			Id:      id,
			EntryID: 0,
			Job:     job,
		}
		callback(id,&myJob)
		jobCache = append(jobCache, myJob)
	}
}


func ChangeJob(id int64,status string,callback Callback)  {
	switch status {
	case "STOP":
		Remove(id)
		break
	case "RUN":
		Run(id,callback)
		break
	case "HOLD":
		Hold(id)
		break
	default:
		break

	}
}
