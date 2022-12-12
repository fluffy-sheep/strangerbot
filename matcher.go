package main

import (
	"context"
	"fmt"
	"runtime"
	"strangerbot/repository"
	"strangerbot/repository/model"
	"strangerbot/service"
	"strangerbot/vars"
	"strings"
	"time"
)

func loopMatchUsers() {

	for {
		matchUsers()
		time.Sleep(1 * time.Second)
		runtime.GC()
		time.Sleep(vars.MatchBackoff)
	}

}

func matchUsers() {

	ctx := context.TODO()

	mud, err := service.ServiceGlobalMatch(ctx)
	if err != nil {
		return
	}

	if len(mud) == 0 {
		return
	}

	repo := repository.GetRepository()

	// find all question
	questions, err := repo.GetAllQuestion(ctx)
	if err != nil {
		return
	}

	// find all question option
	options, err := repo.GetAllOption(ctx)
	if err != nil {
		return
	}

	for i, item := range mud {

		// save match
		if err := service.ServiceSaveMatch(ctx, mud[i]); err != nil {
			continue
		}

		// find user all user question data
		userQuestionData, err := repo.GetUserQuestionDataByUser(ctx, item.User.ChatID)
		if err != nil {
			continue
		}

		matchedUserQuestionData, err := repo.GetUserQuestionDataByUser(ctx, item.MatchMatchUserData.User.ChatID)
		if err != nil {
			continue
		}

		// user and matched user profile
		profileQuestion := model.Questions(questions).GetProfileQuestion()
		profileOptions := model.Options(options).GetQuestionOptions(ctx, profileQuestion)
		userProfile := model.UserQuestionDataList(userQuestionData).GetUserQuestionDataByOptions(ctx, profileOptions)
		matchedUserProfile := model.UserQuestionDataList(matchedUserQuestionData).GetUserQuestionDataByOptions(ctx, profileOptions)
		userProfileStr := strings.Join(profileOptions.GetOptionsByIds(userProfile.GetOptionIds()), ",")
		matchedUserProfileStr := strings.Join(profileOptions.GetOptionsByIds(matchedUserProfile.GetOptionIds()), ",")

		// user and matched user goals
		var userGoals, matchedUserGoals string

		if vars.GoalsQuestionId > 0 {
			goalsOptions := model.Options(options).GetOptionsByQuestionId(vars.GoalsQuestionId)
			userGoalsOptions := model.UserQuestionDataList(userQuestionData).GetUserQuestionDataByOptions(ctx, goalsOptions)
			userGoals = strings.Join(model.Options(options).GetOptionsByIds(userGoalsOptions.GetOptionIds()), ",")
			matchedUserGoalsOptions := model.UserQuestionDataList(matchedUserQuestionData).GetUserQuestionDataByOptions(ctx, goalsOptions)
			matchedUserGoals = strings.Join(model.Options(options).GetOptionsByIds(matchedUserGoalsOptions.GetOptionIds()), ",")
		}

		sendMatchMessage(item.User.ChatID, item.MatchMatchUserData.User.ChatID, userProfileStr, userGoals, matchedUserProfileStr, matchedUserGoals)

	}

}

func sendMatchMessage(userChatId, matchUserChatId int64, userProfile, userGoals, matchedUserProfile, matchedUserGoals string) {

	if len(userProfile) == 0 {
		userProfile = "(!NOT SETTING)"
	}

	if len(userGoals) == 0 {
		userGoals = "(!NOT SETTING)"
	}

	if len(matchedUserProfile) == 0 {
		matchedUserProfile = "(!NOT SETTING)"
	}

	if len(matchedUserGoals) == 0 {
		matchedUserGoals = "(!NOT SETTING)"
	}

	_, _ = telegram.SendMessage(matchUserChatId, fmt.Sprintf(vars.MatchedMessage, userProfile, userGoals), emptyOpts)
	_, _ = telegram.SendMessage(userChatId, fmt.Sprintf(vars.MatchedMessage, matchedUserProfile, matchedUserGoals), emptyOpts)

}
